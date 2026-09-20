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

// quantityEpsilon 余量浮点比较容差：扣减后低于该值视为归零。
const quantityEpsilon = 1e-9

// DisposalRequestService 临期处置申请服务：提交、审批、驳回闭环。
type DisposalRequestService struct {
	repo        *repository.DisposalRequestRepository
	foodRepo    *repository.FoodItemRepository
	consumeRepo *repository.ConsumptionRecordRepository
	familySvc   *FamilyGroupService
	calculator  *util.FoodCalculator
	log         *slog.Logger
}

// NewDisposalRequestService 构造处置申请服务。
func NewDisposalRequestService(
	repo *repository.DisposalRequestRepository,
	foodRepo *repository.FoodItemRepository,
	consumeRepo *repository.ConsumptionRecordRepository,
	familySvc *FamilyGroupService,
	calculator *util.FoodCalculator,
	log *slog.Logger,
) *DisposalRequestService {
	return &DisposalRequestService{
		repo: repo, foodRepo: foodRepo, consumeRepo: consumeRepo,
		familySvc: familySvc, calculator: calculator, log: log,
	}
}

// CreateDisposalInput 提交处置申请入参。
type CreateDisposalInput struct {
	FoodItemID uint    `json:"food_item_id"`
	Quantity   float64 `json:"quantity"`
	Method     string  `json:"method"`
	Reason     string  `json:"reason"`
}

// Create 家庭成员提交临期/过期食品处置申请。
func (s *DisposalRequestService) Create(ctx context.Context, applicantID uint, input CreateDisposalInput) (*model.DisposalRequest, error) {
	item, err := s.foodRepo.FindByID(input.FoodItemID)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError("食品（FoodItem）不存在", err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_CREATE_FAILED, fmt.Errorf("find food item: %w", err))
	}
	if err := s.familySvc.IsMember(ctx, item.FamilyID, applicantID); err != nil {
		return nil, err
	}
	if !contains(constants.DisposalMethods, input.Method) {
		return nil, util.BadRequest(constants.MsgDisposalMethodInvalid, errors.New("invalid disposal method"))
	}
	if input.Quantity <= 0 {
		return nil, util.BadRequest("处置数量（DisposalRequest.quantity）必须大于 0", errors.New("quantity must be positive"))
	}
	// 仅临期/过期且未消耗的食品可申请。
	currentStatus := s.calculator.ComputeFreshness(item.Status, item.ExpiryDate)
	if currentStatus == constants.FreshnessConsumed {
		return nil, util.NewAppError(constants.CodeFoodNotAvailable, 409, constants.MsgDisposalFoodConsumed,
			errors.New("food already consumed"))
	}
	if currentStatus != constants.FreshnessExpiring && currentStatus != constants.FreshnessExpired {
		return nil, util.NewAppError(constants.CodeDisposalNotEligible, 409, constants.MsgDisposalNotEligible,
			errors.New("food is not expiring or expired"))
	}
	// 超量整单拒绝（提交时预校验，审批时在行锁下再次校验）。
	if input.Quantity > item.Quantity+quantityEpsilon {
		return nil, util.NewAppError(constants.CodeQuantityExceeds, 409,
			fmt.Sprintf("FoodItem[id=%d] disposal request rejected: quantity %.3f exceeds stock %.3f", item.ID, input.Quantity, item.Quantity),
			util.ErrQuantityExceeds)
	}
	// 同一食品只允许一张待处理申请。
	if existing, err := s.repo.FindPendingByFood(input.FoodItemID); err == nil && existing != nil {
		return nil, util.NewAppError(constants.CodeDisposalDuplicate, 409, constants.MsgDisposalDuplicate,
			errors.New("pending disposal request already exists"))
	} else if err != nil && !errors.Is(err, util.ErrNotFound) {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_CREATE_FAILED, fmt.Errorf("check pending request: %w", err))
	}

	req := &model.DisposalRequest{
		FamilyID:    item.FamilyID,
		FoodItemID:  input.FoodItemID,
		ApplicantID: applicantID,
		Quantity:    input.Quantity,
		Method:      input.Method,
		Reason:      input.Reason,
		Status:      constants.DisposalStatusPending,
	}
	if err := s.repo.Create(req); err != nil {
		// 并发提交时由部分唯一索引兜底：整单拒绝，不产生任何库存变更。
		if repository.IsDuplicateKeyErr(err) {
			return nil, util.NewAppError(constants.CodeDisposalDuplicate, 409, constants.MsgDisposalDuplicate, err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_CREATE_FAILED, fmt.Errorf("create disposal request: %w", err))
	}
	created, err := s.repo.FindByID(req.ID)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_CREATE_FAILED, fmt.Errorf("load disposal request: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_DISPOSAL_CREATED,
		"disposal_id", req.ID, "food_id", input.FoodItemID, "applicant_id", applicantID,
		"quantity", input.Quantity, "method", input.Method)
	return created, nil
}

// Approve 家庭管理员批准申请：行锁 + CAS 保证并发仅一次成功；扣减余量、生成消耗记录并结案。
func (s *DisposalRequestService) Approve(ctx context.Context, reviewerID, id uint) (*model.DisposalRequest, error) {
	req, err := s.loadForReview(ctx, reviewerID, id)
	if err != nil {
		return nil, err
	}
	if req.Status != constants.DisposalStatusPending {
		return nil, util.NewAppError(constants.CodeDisposalNotPending, 409, constants.MsgDisposalNotPending,
			errors.New("disposal request already closed"))
	}
	s.log.InfoContext(ctx, constants.LOG_DISPOSAL_APPROVE_START, "disposal_id", id, "reviewer_id", reviewerID)

	now := time.Now()
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		txReqRepo := s.repo.WithTx(tx)
		txFoodRepo := s.foodRepo.WithTx(tx)
		txConsumeRepo := s.consumeRepo.WithTx(tx)

		// 1. 锁定申请行并复核状态（PG 行锁串行化并发批准；CAS 结案再做最终保证）。
		locked, err := txReqRepo.FindByIDForUpdate(id)
		if err != nil {
			return fmt.Errorf("lock disposal request: %w", err)
		}
		if locked.Status != constants.DisposalStatusPending {
			return util.NewAppError(constants.CodeDisposalNotPending, 409, constants.MsgDisposalNotPending,
				errors.New("disposal request already closed"))
		}

		// 2. 行锁读取食品并复核：超量或已消耗整单拒绝（回滚，库存不变）。
		item, err := txFoodRepo.FindByIDForUpdate(req.FoodItemID)
		if err != nil {
			return fmt.Errorf("lock food item: %w", err)
		}
		if item.Status == constants.FreshnessConsumed || item.Quantity <= quantityEpsilon {
			return util.NewAppError(constants.CodeFoodNotAvailable, 409,
				fmt.Sprintf("FoodItem[id=%d] disposal approve rejected: food already consumed", item.ID),
				errors.New("food already consumed"))
		}
		if req.Quantity > item.Quantity+quantityEpsilon {
			return util.NewAppError(constants.CodeQuantityExceeds, 409,
				fmt.Sprintf("FoodItem[id=%d] disposal approve rejected: quantity %.3f exceeds stock %.3f", item.ID, req.Quantity, item.Quantity),
				util.ErrQuantityExceeds)
		}

		// 3. 条件原子扣减余量（WHERE quantity >= ?）；归零自动标记已消耗。
		// 与步骤 2 同一事务，0 行说明并发下库存已变，整单回滚。
		rows, err := txFoodRepo.DeductQuantity(item.ID, req.Quantity, now)
		if err != nil {
			return fmt.Errorf("deduct food quantity: %w", err)
		}
		if rows == 0 {
			return util.NewAppError(constants.CodeQuantityExceeds, 409,
				fmt.Sprintf("FoodItem[id=%d] disposal approve rejected: stock changed concurrently", item.ID),
				util.ErrQuantityExceeds)
		}

		// 4. 生成消耗记录（处置方式对应的数量消耗，操作人为批准的家庭管理员）。
		record := &model.ConsumptionRecord{
			FoodItemID: item.ID,
			Quantity:   req.Quantity,
			UserID:     reviewerID,
			ConsumedAt: now,
		}
		if err := txConsumeRepo.Create(record); err != nil {
			return fmt.Errorf("create consumption record: %w", err)
		}

		// 5. CAS 结案：并发批准下最多一条事务生效；失败者整体回滚（库存恢复、记录撤销）。
		closed, err := txReqRepo.Approve(id, reviewerID, record.ID, now)
		if err != nil {
			return fmt.Errorf("approve disposal request: %w", err)
		}
		if closed == 0 {
			return util.NewAppError(constants.CodeDisposalNotPending, 409, constants.MsgDisposalNotPending,
				errors.New("disposal request concurrently closed"))
		}
		return nil
	})
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_APPROVE_FAILED, err)
	}

	approved, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_APPROVED, fmt.Errorf("load approved request: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_DISPOSAL_APPROVED,
		"disposal_id", id, "food_id", req.FoodItemID, "quantity", req.Quantity, "reviewer_id", reviewerID)
	return approved, nil
}

// Reject 家庭管理员驳回申请：仅关闭申请，不改库存。
func (s *DisposalRequestService) Reject(ctx context.Context, reviewerID, id uint, remark string) (*model.DisposalRequest, error) {
	if _, err := s.loadForReview(ctx, reviewerID, id); err != nil {
		return nil, err
	}
	rows, err := s.repo.Reject(id, reviewerID, remark, time.Now())
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_REJECT_FAILED, fmt.Errorf("reject disposal request: %w", err))
	}
	if rows == 0 {
		return nil, util.NewAppError(constants.CodeDisposalNotPending, 409, constants.MsgDisposalNotPending,
			errors.New("disposal request already closed"))
	}
	rejected, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_REJECTED, fmt.Errorf("load rejected request: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_DISPOSAL_REJECTED, "disposal_id", id, "reviewer_id", reviewerID)
	return rejected, nil
}

// loadForReview 加载申请并校验审批人是该家庭管理员。
func (s *DisposalRequestService) loadForReview(ctx context.Context, reviewerID, id uint) (*model.DisposalRequest, error) {
	req, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgDisposalNotFound, err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_APPROVE_FAILED, fmt.Errorf("find disposal request: %w", err))
	}
	if _, err := s.familySvc.RequireAdmin(ctx, req.FamilyID, reviewerID); err != nil {
		return nil, err
	}
	return req, nil
}

// List 分页查询家庭处置申请（可按食品/状态筛选）。
func (s *DisposalRequestService) List(ctx context.Context, userID, familyID, foodItemID uint, status string, page, pageSize int) ([]model.DisposalRequest, int64, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return nil, 0, err
	}
	if status != "" && !contains(constants.DisposalStatuses, status) {
		return nil, 0, util.BadRequest("处置申请状态（DisposalRequest.status）不合法", errors.New("invalid status"))
	}
	list, total, err := s.repo.List(familyID, foodItemID, status, page, pageSize)
	if err != nil {
		return nil, 0, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_LISTED, fmt.Errorf("list disposal requests: %w", err))
	}
	return list, total, nil
}

// GetByID 查询申请详情（需为同家庭成员）。
func (s *DisposalRequestService) GetByID(ctx context.Context, userID, id uint) (*model.DisposalRequest, error) {
	req, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgDisposalNotFound, err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_DISPOSAL_LISTED, fmt.Errorf("find disposal request: %w", err))
	}
	if err := s.familySvc.IsMember(ctx, req.FamilyID, userID); err != nil {
		return nil, err
	}
	return req, nil
}
