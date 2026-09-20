package repository

import (
	"errors"

	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// FamilyGroupRepository 家庭组仓储。
type FamilyGroupRepository struct{ db *gorm.DB }

// NewFamilyGroupRepository 构造家庭组仓储。
func NewFamilyGroupRepository(db *gorm.DB) *FamilyGroupRepository { return &FamilyGroupRepository{db: db} }

// WithTx 使用事务连接构造仓储，便于 service 层在事务内编排多次写入。
func (r *FamilyGroupRepository) WithTx(tx *gorm.DB) *FamilyGroupRepository {
	return &FamilyGroupRepository{db: tx}
}

// Transaction 在事务内执行 fn，任一步返回 error 则整体回滚。
func (r *FamilyGroupRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Create 创建家庭组。
func (r *FamilyGroupRepository) Create(group *model.FamilyGroup) error { return r.db.Create(group).Error }

// FindByID 按 ID 查询（含 Owner）。
func (r *FamilyGroupRepository) FindByID(id uint) (*model.FamilyGroup, error) {
	var group model.FamilyGroup
	if err := r.db.Preload("Owner").First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &group, nil
}

// FindByInviteCode 按邀请码查询。
func (r *FamilyGroupRepository) FindByInviteCode(code string) (*model.FamilyGroup, error) {
	var group model.FamilyGroup
	if err := r.db.Where("invite_code = ?", code).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &group, nil
}

// ListByOwner 按 owner 列出。
func (r *FamilyGroupRepository) ListByOwner(ownerID uint) ([]model.FamilyGroup, error) {
	var groups []model.FamilyGroup
	err := r.db.Preload("Owner").Where("owner_id = ?", ownerID).Find(&groups).Error
	return groups, err
}

// ListByIDs 按 ID 列表查询。
func (r *FamilyGroupRepository) ListByIDs(ids []uint) ([]model.FamilyGroup, error) {
	var groups []model.FamilyGroup
	if len(ids) == 0 {
		return groups, nil
	}
	err := r.db.Preload("Owner").Where("id IN ?", ids).Find(&groups).Error
	return groups, err
}

// Update 更新家庭组。
func (r *FamilyGroupRepository) Update(group *model.FamilyGroup) error { return r.db.Save(group).Error }
