package repository

import (
	"errors"

	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// RecipeRepository 食谱仓储。
type RecipeRepository struct{ db *gorm.DB }

// NewRecipeRepository 构造食谱仓储。
func NewRecipeRepository(db *gorm.DB) *RecipeRepository { return &RecipeRepository{db: db} }

// Create 创建食谱。
func (r *RecipeRepository) Create(recipe *model.Recipe) error { return r.db.Create(recipe).Error }

// FindByID 按 ID 查询。
func (r *RecipeRepository) FindByID(id uint) (*model.Recipe, error) {
	var recipe model.Recipe
	if err := r.db.First(&recipe, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &recipe, nil
}

// ListByCategories 按适用类别查询食谱。
func (r *RecipeRepository) ListByCategories(categories []string) ([]model.Recipe, error) {
	var recipes []model.Recipe
	if len(categories) == 0 {
		err := r.db.Order("id asc").Find(&recipes).Error
		return recipes, err
	}
	err := r.db.Where("suitable_category IN ?", categories).Order("id asc").Find(&recipes).Error
	return recipes, err
}

// ListAll 查询全部食谱。
func (r *RecipeRepository) ListAll() ([]model.Recipe, error) {
	var recipes []model.Recipe
	err := r.db.Order("id asc").Find(&recipes).Error
	return recipes, err
}
