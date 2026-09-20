package repository

import (
	"errors"

	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// DisposalApplicationRepository 临期处置申请仓储。
type DisposalApplicationRepository struct{ db *gorm.DB }

// NewDisposalApplicationRepository 构造处置申请仓储。
func NewDisposalApplicationRepository(db *gorm.DB) *DisposalApplicationRepository {
	return &DisposalApplicationRepository{db: db}
}

// WithTx 使用事务连接构造仓储。
func (r *DisposalApplicationRepository) WithTx(tx *gorm.DB) *DisposalApplicationRepository {
	return &DisposalApplicationRepository{db: tx}
}

// Transaction 在事务内执行 fn，任一步返回 error 则整体回滚。
func (r *DisposalApplicationRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Create 创建处置申请。
func (r *DisposalApplicationRepository) Create(app *model.DisposalApplication) error {
	return r.db.Create(app).Error
}

// FindByID 按 ID 查询（含申请人、审批人、食品）。
func (r *DisposalApplicationRepository) FindByID(id uint) (*model.DisposalApplication, error) {
	var app model.DisposalApplication
	err := r.db.Preload("Applicant").Preload("Reviewer").Preload("FoodItem").First(&app, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &app, nil
}

// ExistsPendingByFood 判断某食品是否已存在待处理申请（同一食品仅允许一张 pending）。
func (r *DisposalApplicationRepository) ExistsPendingByFood(foodItemID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.DisposalApplication{}).
		Where("food_item_id = ? AND status = ?", foodItemID, "pending").
		Count(&count).Error
	return count > 0, err
}

// List 按家庭/状态/食品分页查询（含申请人、审批人、食品）。
func (r *DisposalApplicationRepository) List(familyID uint, status string, foodItemID uint, page, pageSize int) ([]model.DisposalApplication, int64, error) {
	filter := func(db *gorm.DB) *gorm.DB {
		q := db.Where("disposal_applications.family_id = ?", familyID)
		if status != "" {
			q = q.Where("disposal_applications.status = ?", status)
		}
		if foodItemID > 0 {
			q = q.Where("disposal_applications.food_item_id = ?", foodItemID)
		}
		return q
	}
	var total int64
	if err := filter(r.db.Model(&model.DisposalApplication{})).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var apps []model.DisposalApplication
	err := r.db.Preload("Applicant").Preload("Reviewer").Preload("FoodItem").
		Scopes(filter).
		Order("disposal_applications.created_at desc").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&apps).Error
	return apps, total, err
}

// ClaimPending 原子抢占待处理申请：仅当 status=pending 时更新为目标状态。
// 并发审批只有一个事务 RowsAffected=1，其余得到 0 → 审批失败且不改库存。
func (r *DisposalApplicationRepository) ClaimPending(id, reviewerID uint, target, note string, reviewedAt any) (int64, error) {
	res := r.db.Model(&model.DisposalApplication{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(map[string]any{
			"status":      target,
			"reviewer_id": reviewerID,
			"review_note": note,
			"reviewed_at": reviewedAt,
		})
	return res.RowsAffected, res.Error
}
