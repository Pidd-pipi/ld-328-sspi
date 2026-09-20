package repository

import (
	"errors"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// FoodItemRepository 食品仓储。
type FoodItemRepository struct{ db *gorm.DB }

// NewFoodItemRepository 构造食品仓储。
func NewFoodItemRepository(db *gorm.DB) *FoodItemRepository { return &FoodItemRepository{db: db} }

// WithTx 使用事务连接构造仓储。
func (r *FoodItemRepository) WithTx(tx *gorm.DB) *FoodItemRepository {
	return &FoodItemRepository{db: tx}
}

// Transaction 在事务内执行 fn，任一步返回 error 则整体回滚。
func (r *FoodItemRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Create 创建食品。
func (r *FoodItemRepository) Create(item *model.FoodItem) error { return r.db.Create(item).Error }

// FindByID 按 ID 查询。
func (r *FoodItemRepository) FindByID(id uint) (*model.FoodItem, error) {
	var item model.FoodItem
	if err := r.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// List 按条件分页查询（家庭/类别/状态/存放位置/关键词）。
func (r *FoodItemRepository) List(familyID uint, category, status, storageLocation, keyword string, page, pageSize int) ([]model.FoodItem, int64, error) {
	q := r.db.Model(&model.FoodItem{}).Where("family_id = ?", familyID)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if storageLocation != "" {
		q = q.Where("storage_location = ?", storageLocation)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name ILIKE ?", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.FoodItem
	err := q.Order("expiry_date asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

// ListByStatus 按状态查询（看板与临期扫描）。
func (r *FoodItemRepository) ListByStatus(familyID uint, statuses []string) ([]model.FoodItem, error) {
	var items []model.FoodItem
	q := r.db.Model(&model.FoodItem{})
	if familyID > 0 {
		q = q.Where("family_id = ?", familyID)
	}
	err := q.Where("status IN ?", statuses).Find(&items).Error
	return items, err
}

// Update 更新食品。
func (r *FoodItemRepository) Update(item *model.FoodItem) error { return r.db.Save(item).Error }

// UpdateStatus 更新新鲜度状态。
func (r *FoodItemRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.FoodItem{}).Where("id = ?", id).Update("status", status).Error
}

// DeductQuantity 原子条件扣减余量（处置批准专用）：
// 仅当食品未消耗且余量充足时扣减；余量归零同时把状态置为 consumed。
// 返回 RowsAffected：0 表示超量或已消耗（并发批准失败方不得改库存）。
func (r *FoodItemRepository) DeductQuantity(id uint, quantity float64) (int64, error) {
	res := r.db.Model(&model.FoodItem{}).
		Where("id = ? AND status <> ? AND quantity >= ?", id, constants.FreshnessConsumed, quantity).
		Updates(map[string]any{
			"quantity": gorm.Expr("quantity - ?", quantity),
			"status": gorm.Expr(
				"CASE WHEN quantity - ? <= 0 THEN ? ELSE status END",
				quantity, constants.FreshnessConsumed,
			),
		})
	return res.RowsAffected, res.Error
}

// Delete 删除食品。
func (r *FoodItemRepository) Delete(id uint) error { return r.db.Delete(&model.FoodItem{}, id).Error }

// CountByFamily 统计家庭食品数。
func (r *FoodItemRepository) CountByFamily(familyID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.FoodItem{}).Where("family_id = ?", familyID).Count(&count).Error
	return count, err
}

// CountGroupByCategory 按类别统计（统计报表）。
func (r *FoodItemRepository) CountGroupByCategory(familyID uint) ([]model.CategoryCount, error) {
	var rows []model.CategoryCount
	err := r.db.Model(&model.FoodItem{}).
		Select("category, count(*) as count, COALESCE(sum(quantity),0) as total_quantity").
		Where("family_id = ?", familyID).Group("category").Scan(&rows).Error
	return rows, err
}

// SumQuantityByCategory 按类别汇总数量（消费报表）。
func (r *FoodItemRepository) SumQuantityByCategory(familyID uint, status string) ([]model.CategoryCount, error) {
	var rows []model.CategoryCount
	q := r.db.Model(&model.FoodItem{}).
		Select("category, count(*) as count, COALESCE(sum(quantity),0) as total_quantity").
		Where("family_id = ?", familyID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Group("category").Scan(&rows).Error
	return rows, err
}
