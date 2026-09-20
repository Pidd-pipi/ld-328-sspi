package repository

import (
	"errors"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DisposalRequestRepository 临期处置申请仓储。
type DisposalRequestRepository struct{ db *gorm.DB }

// NewDisposalRequestRepository 构造处置申请仓储。
func NewDisposalRequestRepository(db *gorm.DB) *DisposalRequestRepository {
	return &DisposalRequestRepository{db: db}
}

// WithTx 使用事务连接构造仓储。
func (r *DisposalRequestRepository) WithTx(tx *gorm.DB) *DisposalRequestRepository {
	return &DisposalRequestRepository{db: tx}
}

// Transaction 在事务内执行 fn，任一步返回 error 则整体回滚。
func (r *DisposalRequestRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Create 创建处置申请。
func (r *DisposalRequestRepository) Create(req *model.DisposalRequest) error {
	return r.db.Create(req).Error
}

// FindByID 按 ID 查询（含食品、申请人、审批人、消耗记录关联）。
func (r *DisposalRequestRepository) FindByID(id uint) (*model.DisposalRequest, error) {
	var req model.DisposalRequest
	if err := r.db.
		Preload("FoodItem").
		Preload("Applicant").
		Preload("Reviewer").
		Preload("ConsumptionRecord").
		First(&req, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &req, nil
}

// FindByIDForUpdate 按 ID 查询申请并加行锁（PostgreSQL FOR UPDATE），必须在事务内调用。
// SQLite 等方言退化为普通查询（测试环境），并发安全由 CAS 结案与条件扣减兜底。
func (r *DisposalRequestRepository) FindByIDForUpdate(id uint) (*model.DisposalRequest, error) {
	var req model.DisposalRequest
	query := r.db
	if r.db.Dialector.Name() == "postgres" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.First(&req, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &req, nil
}

// FindPendingByFood 查询某食品当前待处理申请；无则返回 util.ErrNotFound。
func (r *DisposalRequestRepository) FindPendingByFood(foodItemID uint) (*model.DisposalRequest, error) {
	var req model.DisposalRequest
	err := r.db.Where("food_item_id = ? AND status = ?", foodItemID, constants.DisposalStatusPending).First(&req).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &req, nil
}

// List 按家庭/食品/状态分页查询处置申请。
func (r *DisposalRequestRepository) List(familyID, foodItemID uint, status string, page, pageSize int) ([]model.DisposalRequest, int64, error) {
	q := r.db.Model(&model.DisposalRequest{}).Where("family_id = ?", familyID)
	if foodItemID > 0 {
		q = q.Where("food_item_id = ?", foodItemID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.DisposalRequest
	err := q.Preload("FoodItem").Preload("Applicant").Preload("Reviewer").Preload("ConsumptionRecord").
		Order("created_at desc").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Approve CAS 结案：仅当申请仍为 pending 时写入批准结果，返回是否抢到本次结案。
// 并发批准时最多一行受影响，失败者不产生任何库存变更（与库存扣减同一事务）。
func (r *DisposalRequestRepository) Approve(id, reviewerID, consumptionRecordID uint, reviewedAt interface{}) (int64, error) {
	res := r.db.Model(&model.DisposalRequest{}).
		Where("id = ? AND status = ?", id, constants.DisposalStatusPending).
		Updates(map[string]interface{}{
			"status":                constants.DisposalStatusApproved,
			"reviewer_id":           reviewerID,
			"reviewed_at":           reviewedAt,
			"consumption_record_id": consumptionRecordID,
		})
	return res.RowsAffected, res.Error
}

// Reject CAS 结案：仅当申请仍为 pending 时驳回，返回是否生效。
func (r *DisposalRequestRepository) Reject(id, reviewerID uint, remark string, reviewedAt interface{}) (int64, error) {
	res := r.db.Model(&model.DisposalRequest{}).
		Where("id = ? AND status = ?", id, constants.DisposalStatusPending).
		Updates(map[string]interface{}{
			"status":        constants.DisposalStatusRejected,
			"reviewer_id":   reviewerID,
			"review_remark": remark,
			"reviewed_at":   reviewedAt,
		})
	return res.RowsAffected, res.Error
}

// IsDuplicateKeyErr 判断是否唯一约束冲突（同一食品重复待处理申请）。
func IsDuplicateKeyErr(err error) bool {
	return isPGDuplicateKey(err) || isSQLiteDuplicateKey(err)
}
