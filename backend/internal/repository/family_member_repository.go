package repository

import (
	"errors"

	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// FamilyMemberRepository 家庭成员仓储。
type FamilyMemberRepository struct{ db *gorm.DB }

// NewFamilyMemberRepository 构造家庭成员仓储。
func NewFamilyMemberRepository(db *gorm.DB) *FamilyMemberRepository {
	return &FamilyMemberRepository{db: db}
}

// WithTx 使用事务连接构造仓储。
func (r *FamilyMemberRepository) WithTx(tx *gorm.DB) *FamilyMemberRepository {
	return &FamilyMemberRepository{db: tx}
}

// Create 添加成员。
func (r *FamilyMemberRepository) Create(member *model.FamilyMember) error {
	return r.db.Create(member).Error
}

// FindByFamilyAndUser 按家庭与用户查询。
func (r *FamilyMemberRepository) FindByFamilyAndUser(familyID, userID uint) (*model.FamilyMember, error) {
	var member model.FamilyMember
	err := r.db.Where("family_id = ? AND user_id = ?", familyID, userID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &member, nil
}

// FindByID 按 ID 查询。
func (r *FamilyMemberRepository) FindByID(id uint) (*model.FamilyMember, error) {
	var member model.FamilyMember
	if err := r.db.Preload("User").First(&member, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &member, nil
}

// ListByFamily 列出家庭成员（含 User）。
func (r *FamilyMemberRepository) ListByFamily(familyID uint) ([]model.FamilyMember, error) {
	var members []model.FamilyMember
	err := r.db.Preload("User").Where("family_id = ?", familyID).Order("joined_at asc").Find(&members).Error
	return members, err
}

// ListByUser 列出用户加入的家庭。
func (r *FamilyMemberRepository) ListByUser(userID uint) ([]model.FamilyMember, error) {
	var members []model.FamilyMember
	err := r.db.Preload("Family").Preload("Family.Owner").Where("user_id = ?", userID).Find(&members).Error
	return members, err
}

// UpdateRole 更新成员角色。
func (r *FamilyMemberRepository) UpdateRole(member *model.FamilyMember, role string) error {
	return r.db.Model(member).Update("role", role).Error
}

// Delete 删除成员。
func (r *FamilyMemberRepository) Delete(member *model.FamilyMember) error {
	return r.db.Delete(member).Error
}

// CountByFamily 统计家庭人数。
func (r *FamilyMemberRepository) CountByFamily(familyID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.FamilyMember{}).Where("family_id = ?", familyID).Count(&count).Error
	return count, err
}
