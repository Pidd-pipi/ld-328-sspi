package repository

import (
	"errors"

	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// UserRepository 用户仓储。
type UserRepository struct{ db *gorm.DB }

// NewUserRepository 构造用户仓储。
func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

// Create 创建用户。
func (r *UserRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// FindByPhone 按手机号查询。
func (r *UserRepository) FindByPhone(phone string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByID 按 ID 查询。
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户。
func (r *UserRepository) Update(user *model.User) error { return r.db.Save(user).Error }

// CountByPhone 按手机号统计（注册查重）。
func (r *UserRepository) CountByPhone(phone string) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("phone = ?", phone).Count(&count).Error
	return count, err
}
