package repository

import (
	"errors"

	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// NotificationRepository 通知仓储。
type NotificationRepository struct{ db *gorm.DB }

// NewNotificationRepository 构造通知仓储。
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// WithTx 使用事务连接构造仓储。
func (r *NotificationRepository) WithTx(tx *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: tx}
}

// Transaction 在事务内执行 fn，任一步返回 error 则整体回滚。
func (r *NotificationRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Create 创建通知。
func (r *NotificationRepository) Create(n *model.Notification) error { return r.db.Create(n).Error }

// CreateBatch 批量创建通知。
func (r *NotificationRepository) CreateBatch(items []model.Notification) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

// FindByID 按 ID 查询。
func (r *NotificationRepository) FindByID(id uint) (*model.Notification, error) {
	var n model.Notification
	if err := r.db.First(&n, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &n, nil
}

// ListByFamily 按家庭查询通知。
func (r *NotificationRepository) ListByFamily(familyID uint, unreadOnly bool, page, pageSize int) ([]model.Notification, int64, error) {
	q := r.db.Model(&model.Notification{}).Where("family_id = ?", familyID)
	if unreadOnly {
		q = q.Where("is_read = ?", false)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Notification
	err := r.db.Preload("FoodItem").Where("family_id = ?", familyID).
		Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

// MarkRead 标记已读。
func (r *NotificationRepository) MarkRead(id uint) error {
	return r.db.Model(&model.Notification{}).Where("id = ?", id).Updates(map[string]any{
		"is_read": true, "read_at": gorm.Expr("now()"),
	}).Error
}

// MarkAllReadByFamily 标记家庭全部已读。
func (r *NotificationRepository) MarkAllReadByFamily(familyID uint) error {
	return r.db.Model(&model.Notification{}).Where("family_id = ? AND is_read = ?", familyID, false).
		Updates(map[string]any{"is_read": true, "read_at": gorm.Expr("now()")}).Error
}

// CountUnreadByFamily 统计未读数量。
func (r *NotificationRepository) CountUnreadByFamily(familyID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("family_id = ? AND is_read = ?", familyID, false).Count(&count).Error
	return count, err
}

// HasForFoodAndType 判断某食品某类型是否已存在通知（扫描去重）。
func (r *NotificationRepository) HasForFoodAndType(foodID uint, typ string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("food_item_id = ? AND type = ?", foodID, typ).Count(&count).Error
	return count > 0, err
}
