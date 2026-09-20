package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
)

// RecipeService 食谱服务：推荐优先食用食品与搭配建议。
type RecipeService struct {
	recipeRepo *repository.RecipeRepository
	foodRepo   *repository.FoodItemRepository
	familySvc  *FamilyGroupService
	calculator *util.FoodCalculator
	log        *slog.Logger
}

// NewRecipeService 构造食谱服务。
func NewRecipeService(recipeRepo *repository.RecipeRepository, foodRepo *repository.FoodItemRepository, familySvc *FamilyGroupService, calculator *util.FoodCalculator, log *slog.Logger) *RecipeService {
	return &RecipeService{recipeRepo: recipeRepo, foodRepo: foodRepo, familySvc: familySvc, calculator: calculator, log: log}
}

// Recommendation 智能推荐：即将过期食品优先食用 + 对应类别食谱。
type Recommendation struct {
	FoodItems []model.FoodItem    `json:"food_items"`
	Recipes   []model.Recipe      `json:"recipes"`
	Summary   RecommendationSummary `json:"summary"`
}

// RecommendationSummary 推荐摘要。
type RecommendationSummary struct {
	ExpiringCount int      `json:"expiring_count"`
	ExpiredCount  int      `json:"expired_count"`
	Categories    []string `json:"categories"`
}

// Recommend 生成推荐。
func (s *RecipeService) Recommend(ctx context.Context, userID, familyID uint) (*Recommendation, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return nil, err
	}
	items, err := s.foodRepo.ListByStatus(familyID, []string{constants.FreshnessExpiring, constants.FreshnessExpired})
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_RECIPE_RECOMMENDED, fmt.Errorf("list expiring foods: %w", err))
	}
	for i := range items {
		items[i].Status = s.calculator.ComputeFreshness(items[i].Status, items[i].ExpiryDate)
	}
	// 优先展示临期食品
	expiring := make([]model.FoodItem, 0)
	categories := make([]string, 0)
	seen := map[string]bool{}
	for _, it := range items {
		if it.Status == constants.FreshnessExpiring {
			expiring = append(expiring, it)
		}
		if !seen[it.Category] {
			seen[it.Category] = true
			categories = append(categories, it.Category)
		}
	}
	recipes, err := s.recipeRepo.ListByCategories(categories)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_RECIPE_RECOMMENDED, fmt.Errorf("list recipes: %w", err))
	}
	result := &Recommendation{
		FoodItems: expiring,
		Recipes:   recipes,
		Summary: RecommendationSummary{
			ExpiringCount: len(expiring),
			ExpiredCount:  countStatus(items, constants.FreshnessExpired),
			Categories:    categories,
		},
	}
	s.log.InfoContext(ctx, constants.LOG_RECIPE_RECOMMENDED, "family_id", familyID, "expiring", len(expiring), "recipes", len(recipes))
	return result, nil
}

func countStatus(items []model.FoodItem, status string) int {
	n := 0
	for _, it := range items {
		if it.Status == status {
			n++
		}
	}
	return n
}

// ListAll 全部食谱（供管理/查看）。
func (s *RecipeService) ListAll(ctx context.Context) ([]model.Recipe, error) {
	return s.recipeRepo.ListAll()
}

// GetByID 查询食谱详情。
func (s *RecipeService) GetByID(ctx context.Context, id uint) (*model.Recipe, error) {
	recipe, err := s.recipeRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError("食谱（Recipe）不存在", err)
		}
		return nil, err
	}
	return recipe, nil
}
