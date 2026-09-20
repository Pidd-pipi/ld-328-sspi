package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// DisposalApplicationService 临期处置申请闭环：申请 → 批准/驳回 → 扣减库存、生成消耗记录、结案。
type DisposalApplicationService struct {
	repo        *repository.DisposalApplicationRepository
	foodRepo    *repository.FoodItemRepository
	consumeRepo *repository.ConsumptionRecordRepository
	familySvc   *FamilyGroupService
	calculator  *util.FoodCalculator
	log         *slog.Logger
}

// NewDisposalApplicationService 构造处置申请服务。
func NewDisposalApplicationService(
	repo *repository.DisposalApplicationRepository,
	foodRepo *repository.FoodItemRepository,
	consumeRepo *repository.ConsumptionRecordRepository,
	familySvc *FamilyGroupService,
	calculator *util.FoodCalculator,
	log *slog.Logger,
) *DisposalApplicationService {
	return &DisposalApplicationService{
		repo: repo, foodRepo: foodRepo, consumeRepo: consumeRepo,
		familySvc: familySvc, calculator: calculator, log: log,
	}
}

// CreateDisposalInput 提交处置申请入参。
type CreateDisposalInput struct {
	FoodItemID uint    `json:"food_item_id"`
	Quantity   float64 `json:"quantity"`
	Method     string  `json:"method"`
}

// Create 家庭成员提交临期处置申请。
// 规则：仅临期/过期食品；数量必须大于 0 且不超过余量；同一食品只允许一张待处理申请。
func (s *DisposalApplicationService) Create(ctx context.Context, userID uint, input CreateDisposalInput) (*model.DisposalApplication, error) {
	item, err := s.foodRepo.FindByID(input.FoodItemID)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError("食品（FoodItem）不存在", err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_CREATED, fmt.Errorf("find food item: %w", err))
	}
	if err := s.familySvc.IsMember(ctx, item.FamilyID, userID); err != nil {
		return nil, err
	}
	switch input.Method {
	case constants.DisposalMethodDiscard, constants.DisposalMethodEat, constants.DisposalMethodDonate:
	default:
		return nil, util.BadRequest(constants.MsgDisposalMethodInvalid, errors.New("invalid disposal method"))
	}
	// 实时刷新新鲜度：已消耗或仍充裕的食品不可申请，仅临期/过期可申请。
	freshness := s.calculator.ComputeFreshness(item.Status, item.ExpiryDate)
	if item.Status == constants.FreshnessConsumed || freshness == constants.FreshnessFresh {
		return nil, util.NewAppError(constants.CodeDisposalNotAllowed, 409, constants.MsgDisposalNotAllowed, errors.New("food not expiring or expired"))
	}
	if input.Quantity <= 0 {
		return nil, util.BadRequest(constants.MsgParamInvalid, errors.New("disposal quantity must be positive"))
	}
	if input.Quantity > item.Quantity {
		return nil, util.NewAppError(constants.CodeQuantityExceeds, 409,
			fmt.Sprintf("FoodItem[id=%d] disposal apply failed: quantity %.2f exceeds stock %.2f", item.ID, input.Quantity, item.Quantity),
			util.ErrQuantityExceeds)
	}
	exists, err := s.repo.ExistsPendingByFood(input.FoodItemID)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_CREATED, fmt.Errorf("check pending application: %w", err))
	}
	if exists {
		return nil, util.NewAppError(constants.CodeDisposalDuplicate, 409, constants.MsgDisposalDuplicate, errors.New("pending application exists"))
	}
	app := &model.DisposalApplication{
		FamilyID: item.FamilyID, FoodItemID: input.FoodItemID, ApplicantID: userID,
		Quantity: input.Quantity, Method: input.Method, Status: constants.DisposalStatusPending,
	}
	if err := s.repo.Create(app); err != nil {
		// 并发提交兜底：部分唯一索引拒绝第二张 pending，转为业务冲突错误。
		if exists, _ := s.repo.ExistsPendingByFood(input.FoodItemID); exists {
			return nil, util.NewAppError(constants.CodeDisposalDuplicate, 409, constants.MsgDisposalDuplicate, errors.New("pending application exists"))
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_CREATED, fmt.Errorf("create disposal application: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_DISPOSAL_CREATED,
		"application_id", app.ID, "food_id", app.FoodItemID, "quantity", app.Quantity,
		"method", app.Method, "applicant_id", userID, "family_id", item.FamilyID)
	return s.repo.FindByID(app.ID)
}

// List 分页查询处置申请列表（支持状态/食品筛选，刷新后可回读）。
func (s *DisposalApplicationService) List(ctx context.Context, userID, familyID uint, status string, foodItemID uint, page, pageSize int) ([]model.DisposalApplication, int64, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return nil, 0, err
	}
	if status != "" && !contains(constants.DisposalStatuses, status) {
		return nil, 0, util.BadRequest(constants.MsgParamInvalid, errors.New("invalid disposal status"))
	}
	apps, total, err := s.repo.List(familyID, status, foodItemID, page, pageSize)
	if err != nil {
		return nil, 0, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_LISTED, fmt.Errorf("list disposal applications: %w", err))
	}
	return apps, total, nil
}

// GetByID 查询处置申请详情（家庭内成员可读）。
func (s *DisposalApplicationService) GetByID(ctx context.Context, userID, id uint) (*model.DisposalApplication, error) {
	app, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeDisposalNotFound, 404, constants.MsgDisposalNotFound, err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_LISTED, fmt.Errorf("find disposal application: %w", err))
	}
	if err := s.familySvc.IsMember(ctx, app.FamilyID, userID); err != nil {
		return nil, err
	}
	return app, nil
}

// Approve 家庭管理员批准申请：原子抢占 + 条件扣减 + 消耗记录，全部在同一事务内。
// 并发批准只能成功一次：抢占失败者 / 超量 / 已消耗整单回滚，不改库存。
func (s *DisposalApplicationService) Approve(ctx context.Context, reviewerID, id uint, note string) (*model.DisposalApplication, error) {
	app, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeDisposalNotFound, 404, constants.MsgDisposalNotFound, err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_APPROVE_FAILED, fmt.Errorf("find disposal application: %w", err))
	}
	if _, err := s.familySvc.RequireAdmin(ctx, app.FamilyID, reviewerID); err != nil {
		return nil, err
	}
	if app.Status != constants.DisposalStatusPending {
		return nil, util.NewAppError(constants.CodeDisposalNotPending, 409, constants.MsgDisposalNotPending, errors.New("application already closed"))
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		// 1. 原子抢占待处理申请：并发审批只有一个事务 RowsAffected=1。
		claimed, err := s.repo.WithTx(tx).ClaimPending(id, reviewerID, constants.DisposalStatusApproved, note, time.Now())
		if err != nil {
			return fmt.Errorf("claim disposal application: %w", err)
		}
		if claimed == 0 {
			return util.NewAppError(constants.CodeDisposalNotPending, 409, constants.MsgDisposalNotPending, errors.New("application already closed concurrently"))
		}
		// 2. 原子条件扣减余量：超量或已消耗返回 0，整单回滚（库存与申请状态都不变）。
		affected, err := s.foodRepo.WithTx(tx).DeductQuantity(app.FoodItemID, app.Quantity)
		if err != nil {
			return fmt.Errorf("deduct food quantity: %w", err)
		}
		if affected == 0 {
			var reason string
			current, ferr := s.foodRepo.WithTx(tx).FindByID(app.FoodItemID)
			switch {
			case ferr != nil:
				reason = "stock unavailable"
			case current.Status == constants.FreshnessConsumed:
				reason = constants.MsgDisposalConsumed
			default:
				reason = fmt.Sprintf("%s（余量 %.2f，申请 %.2f）", constants.MsgDisposalQuantity, current.Quantity, app.Quantity)
			}
			return util.NewAppError(constants.CodeQuantityExceeds, 409,
				fmt.Sprintf("DisposalApplication[id=%d] approve rejected: %s", id, reason), errors.New(reason))
		}
		// 3. 生成消耗记录（审批人作为操作人），处置方式记入记录，供历史与频率分析。
		record := &model.ConsumptionRecord{
			FoodItemID: app.FoodItemID, Quantity: app.Quantity,
			UserID: reviewerID, ConsumedAt: time.Now(),
		}
		if err := s.consumeRepo.WithTx(tx).Create(record); err != nil {
			return fmt.Errorf("create consumption record: %w", err)
		}
		return nil
	})
	if err != nil {
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_APPROVE_FAILED, err)
	}

	approved, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_APPROVE_FAILED, fmt.Errorf("reload approved application: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_DISPOSAL_APPROVED,
		"application_id", id, "food_id", approved.FoodItemID, "quantity", approved.Quantity,
		"method", approved.Method, "reviewer_id", reviewerID, "family_id", approved.FamilyID)
	return approved, nil
}

// Reject 家庭管理员驳回申请：仅关闭申请，不扣减余量、不生成消耗记录。
func (s *DisposalApplicationService) Reject(ctx context.Context, reviewerID, id uint, note string) (*model.DisposalApplication, error) {
	app, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeDisposalNotFound, 404, constants.MsgDisposalNotFound, err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_REJECT_FAILED, fmt.Errorf("find disposal application: %w", err))
	}
	if _, err := s.familySvc.RequireAdmin(ctx, app.FamilyID, reviewerID); err != nil {
		return nil, err
	}
	if app.Status != constants.DisposalStatusPending {
		return nil, util.NewAppError(constants.CodeDisposalNotPending, 409, constants.MsgDisposalNotPending, errors.New("application already closed"))
	}
	affected, err := s.repo.ClaimPending(id, reviewerID, constants.DisposalStatusRejected, note, time.Now())
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_REJECT_FAILED, fmt.Errorf("reject disposal application: %w", err))
	}
	if affected == 0 {
		return nil, util.NewAppError(constants.CodeDisposalNotPending, 409, constants.MsgDisposalNotPending, errors.New("application already closed concurrently"))
	}
	rejected, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_REJECT_FAILED, fmt.Errorf("reload rejected application: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_DISPOSAL_REJECTED,
		"application_id", id, "reviewer_id", reviewerID, "note", note, "family_id", rejected.FamilyID)
	return rejected, nil
}
