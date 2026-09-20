package repository

import (
	"github.com/blueship581/cyfreshfood/internal/model"
	"gorm.io/gorm"
)

// ConsumptionRecordRepository 消耗记录仓储。
type ConsumptionRecordRepository struct{ db *gorm.DB }

// NewConsumptionRecordRepository 构造消耗记录仓储。
func NewConsumptionRecordRepository(db *gorm.DB) *ConsumptionRecordRepository {
	return &ConsumptionRecordRepository{db: db}
}

// WithTx 使用事务连接构造仓储。
func (r *ConsumptionRecordRepository) WithTx(tx *gorm.DB) *ConsumptionRecordRepository {
	return &ConsumptionRecordRepository{db: tx}
}

// Create 创建消耗记录。
func (r *ConsumptionRecordRepository) Create(record *model.ConsumptionRecord) error {
	return r.db.Create(record).Error
}

// ListByFamily 按家庭查询消耗记录（含食品与用户）。
func (r *ConsumptionRecordRepository) ListByFamily(familyID uint, page, pageSize int) ([]model.ConsumptionRecord, int64, error) {
	q := r.db.Model(&model.ConsumptionRecord{}).
		Joins("JOIN food_items ON food_items.id = consumption_records.food_item_id").
		Where("food_items.family_id = ?", familyID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []model.ConsumptionRecord
	err := r.db.Preload("User").Preload("FoodItem").
		Joins("JOIN food_items ON food_items.id = consumption_records.food_item_id").
		Where("food_items.family_id = ?", familyID).
		Order("consumption_records.consumed_at desc").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error
	return records, total, err
}

// ListByFood 按食品查询消耗历史。
func (r *ConsumptionRecordRepository) ListByFood(foodItemID uint) ([]model.ConsumptionRecord, error) {
	var records []model.ConsumptionRecord
	err := r.db.Preload("User").Where("food_item_id = ?", foodItemID).Order("consumed_at desc").Find(&records).Error
	return records, err
}

// CountGroupByCategory 按食品类别统计消耗次数与数量。
func (r *ConsumptionRecordRepository) CountGroupByCategory(familyID uint) ([]model.CategoryCount, error) {
	var rows []model.CategoryCount
	err := r.db.Model(&model.ConsumptionRecord{}).
		Select("food_items.category as category, count(consumption_records.id) as count, COALESCE(sum(consumption_records.quantity),0) as total_quantity").
		Joins("JOIN food_items ON food_items.id = consumption_records.food_item_id").
		Where("food_items.family_id = ?", familyID).
		Group("food_items.category").Scan(&rows).Error
	return rows, err
}

// MonthlyStats 按月统计消耗（month 形如 2026-08）。
func (r *ConsumptionRecordRepository) MonthlyStats(familyID uint, month string) ([]model.MonthlyConsumption, error) {
	var rows []model.MonthlyConsumption
	q := r.db.Model(&model.ConsumptionRecord{}).
		Select("food_items.category as category, count(consumption_records.id) as count, COALESCE(sum(consumption_records.quantity),0) as total_quantity").
		Joins("JOIN food_items ON food_items.id = consumption_records.food_item_id").
		Where("food_items.family_id = ?", familyID)
	if month != "" {
		q = q.Where("to_char(consumption_records.consumed_at, 'YYYY-MM') = ?", month)
	}
	err := q.Group("food_items.category").Scan(&rows).Error
	return rows, err
}

// TopConsumedFoods 最常消耗食品 Top N。
func (r *ConsumptionRecordRepository) TopConsumedFoods(familyID uint, month string, limit int) ([]model.TopFood, error) {
	q := r.db.Model(&model.ConsumptionRecord{}).
		Select("food_items.id as food_item_id, food_items.name as name, count(consumption_records.id) as count, COALESCE(sum(consumption_records.quantity),0) as quantity").
		Joins("JOIN food_items ON food_items.id = consumption_records.food_item_id").
		Where("food_items.family_id = ?", familyID)
	if month != "" {
		q = q.Where("to_char(consumption_records.consumed_at, 'YYYY-MM') = ?", month)
	}
	var rows []model.TopFood
	err := q.Group("food_items.id, food_items.name").Order("count desc").Limit(limit).Scan(&rows).Error
	return rows, err
}
