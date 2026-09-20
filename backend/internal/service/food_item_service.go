package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// FoodItemService 食品服务：录入、编辑、消耗、状态流转、CSV 导入。
type FoodItemService struct {
	repo        *repository.FoodItemRepository
	consumeRepo *repository.ConsumptionRecordRepository
	familySvc   *FamilyGroupService
	calculator  *util.FoodCalculator
	log         *slog.Logger
}

// NewFoodItemService 构造食品服务。
func NewFoodItemService(repo *repository.FoodItemRepository, consumeRepo *repository.ConsumptionRecordRepository, familySvc *FamilyGroupService, calculator *util.FoodCalculator, log *slog.Logger) *FoodItemService {
	return &FoodItemService{repo: repo, consumeRepo: consumeRepo, familySvc: familySvc, calculator: calculator, log: log}
}

// CreateFoodInput 创建食品入参。
type CreateFoodInput struct {
	FamilyID        uint       `json:"family_id" binding:"required"`
	Name            string     `json:"name" binding:"required,max=100"`
	Category        string     `json:"category" binding:"required"`
	ProductionDate  *time.Time `json:"production_date"`
	ShelfLifeDays   int        `json:"shelf_life_days"`
	Quantity        float64    `json:"quantity" binding:"gte=0"`
	Unit            string     `json:"unit"`
	StorageLocation string     `json:"storage_location"`
	OpenedAt        *time.Time `json:"opened_at"`
	ImageURL        string     `json:"image_url"`
}

// Create 录入食品并计算到期日与状态。
func (s *FoodItemService) Create(ctx context.Context, userID uint, input CreateFoodInput) (*model.FoodItem, error) {
	if err := s.familySvc.IsMember(ctx, input.FamilyID, userID); err != nil {
		return nil, err
	}
	if !contains(constants.FoodCategories, input.Category) {
		return nil, util.BadRequest(constants.MsgCategoryInvalid, errors.New("invalid category"))
	}
	item := &model.FoodItem{
		FamilyID: input.FamilyID, Name: input.Name, Category: input.Category,
		ProductionDate: input.ProductionDate, ShelfLifeDays: input.ShelfLifeDays,
		Quantity: input.Quantity, Unit: input.Unit, StorageLocation: input.StorageLocation,
		OpenedAt: input.OpenedAt, ImageURL: input.ImageURL, CreatorID: userID,
	}
	item.ExpiryDate = s.calculator.CalculateExpiryDate(item.ProductionDate, item.ShelfLifeDays, item.OpenedAt)
	item.Status = s.calculator.ComputeFreshness("", item.ExpiryDate)
	if err := s.repo.Create(item); err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_FOOD_CREATED, fmt.Errorf("create food item: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_FOOD_CREATED, "food_id", item.ID, "family_id", item.FamilyID, "category", item.Category)
	return item, nil
}

// List 查询食品列表并实时刷新状态。
func (s *FoodItemService) List(ctx context.Context, userID, familyID uint, category, status, storageLocation, keyword string, page, pageSize int) ([]model.FoodItem, int64, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return nil, 0, err
	}
	items, total, err := s.repo.List(familyID, category, status, storageLocation, keyword, page, pageSize)
	if err != nil {
		return nil, 0, util.LogError(s.log, ctx, constants.LOG_FOOD_STATUS_REFRESHED, fmt.Errorf("list food items: %w", err))
	}
	for i := range items {
		items[i].Status = s.calculator.ComputeFreshness(items[i].Status, items[i].ExpiryDate)
	}
	return items, total, nil
}

// GetByID 查询食品详情。
func (s *FoodItemService) GetByID(ctx context.Context, userID, id uint) (*model.FoodItem, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError("食品（FoodItem）不存在", err)
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_FOOD_UPDATED, fmt.Errorf("find food item: %w", err))
	}
	if err := s.familySvc.IsMember(ctx, item.FamilyID, userID); err != nil {
		return nil, err
	}
	item.Status = s.calculator.ComputeFreshness(item.Status, item.ExpiryDate)
	return item, nil
}

// Update 编辑食品。
func (s *FoodItemService) Update(ctx context.Context, userID, id uint, input CreateFoodInput) (*model.FoodItem, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.NotFoundError("食品（FoodItem）不存在", err)
	}
	if err := s.familySvc.IsMember(ctx, item.FamilyID, userID); err != nil {
		return nil, err
	}
	if input.Name != "" {
		item.Name = input.Name
	}
	if input.Category != "" {
		if !contains(constants.FoodCategories, input.Category) {
			return nil, util.BadRequest(constants.MsgCategoryInvalid, errors.New("invalid category"))
		}
		item.Category = input.Category
	}
	if input.ProductionDate != nil {
		item.ProductionDate = input.ProductionDate
	}
	if input.ShelfLifeDays > 0 {
		item.ShelfLifeDays = input.ShelfLifeDays
	}
	if input.Quantity >= 0 {
		item.Quantity = input.Quantity
	}
	if input.Unit != "" {
		item.Unit = input.Unit
	}
	if input.StorageLocation != "" {
		item.StorageLocation = input.StorageLocation
	}
	if input.OpenedAt != nil {
		item.OpenedAt = input.OpenedAt
	}
	if input.ImageURL != "" {
		item.ImageURL = input.ImageURL
	}
	item.ExpiryDate = s.calculator.CalculateExpiryDate(item.ProductionDate, item.ShelfLifeDays, item.OpenedAt)
	item.Status = s.calculator.ComputeFreshness(item.Status, item.ExpiryDate)
	if err := s.repo.Update(item); err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_FOOD_UPDATED, fmt.Errorf("update food item: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_FOOD_UPDATED, "food_id", id)
	return item, nil
}

// Delete 删除食品。
func (s *FoodItemService) Delete(ctx context.Context, userID, id uint) error {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return util.NotFoundError("食品（FoodItem）不存在", err)
	}
	if err := s.familySvc.IsMember(ctx, item.FamilyID, userID); err != nil {
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return util.LogError(s.log, ctx, constants.LOG_FOOD_DELETED, fmt.Errorf("delete food item: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_FOOD_DELETED, "food_id", id)
	return nil
}

// Consume 消耗食品（创建消耗记录并更新余量/状态）。
func (s *FoodItemService) Consume(ctx context.Context, userID, foodID uint, quantity float64, consumedAt time.Time) (*model.ConsumptionRecord, error) {
	item, err := s.repo.FindByID(foodID)
	if err != nil {
		return nil, util.NotFoundError("食品（FoodItem）不存在", err)
	}
	if err := s.familySvc.IsMember(ctx, item.FamilyID, userID); err != nil {
		return nil, err
	}
	if item.Status == constants.FreshnessConsumed {
		return nil, util.NewAppError(constants.CodeFoodNotAvailable, 409, constants.MsgFoodNotAvailable, errors.New("food consumed"))
	}
	if quantity <= 0 || quantity > item.Quantity {
		return nil, util.NewAppError(constants.CodeQuantityExceeds, 409,
			fmt.Sprintf("FoodItem[id=%d] consume failed: quantity exceeds stock", item.ID),
			util.ErrQuantityExceeds)
	}
	item.Quantity -= quantity
	if item.Quantity == 0 {
		item.Status = constants.FreshnessConsumed
	}
	record := &model.ConsumptionRecord{FoodItemID: foodID, Quantity: quantity, UserID: userID, ConsumedAt: consumedAt}
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.WithTx(tx).Update(item); err != nil {
			return fmt.Errorf("update food quantity: %w", err)
		}
		if err := s.consumeRepo.WithTx(tx).Create(record); err != nil {
			return fmt.Errorf("create consumption record: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_FOOD_CONSUME_FAILED, err)
	}
	s.log.InfoContext(ctx, constants.LOG_FOOD_CONSUME_SUCCESS, "food_id", foodID, "quantity", quantity, "user_id", userID)
	return record, nil
}

// ImportCSV CSV 批量导入（name,category,quantity,unit,shelf_life_days,storage_location）。
func (s *FoodItemService) ImportCSV(ctx context.Context, userID, familyID uint, csvText string) (int, []model.FoodItem, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return 0, nil, err
	}
	reader := csv.NewReader(strings.NewReader(csvText))
	created := make([]model.FoodItem, 0, 16)
	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, nil, util.BadRequest("CSV 格式（FoodItem.csv）不合法", err)
		}
		if len(row) < 2 || strings.TrimSpace(row[0]) == "name" || strings.TrimSpace(row[0]) == "" {
			continue
		}
		name := strings.TrimSpace(row[0])
		category := strings.TrimSpace(row[1])
		if category == "" {
			category = constants.FoodCategoryOther
		}
		if !contains(constants.FoodCategories, category) {
			category = constants.FoodCategoryOther
		}
		quantity := 1.0
		unit := "份"
		days := 7
		location := constants.StorageFridge
		if len(row) > 2 {
			if v, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64); err == nil {
				quantity = v
			}
		}
		if len(row) > 3 && strings.TrimSpace(row[3]) != "" {
			unit = strings.TrimSpace(row[3])
		}
		if len(row) > 4 {
			if v, err := strconv.Atoi(strings.TrimSpace(row[4])); err == nil && v > 0 {
				days = v
			}
		}
		if len(row) > 5 && strings.TrimSpace(row[5]) != "" {
			location = strings.TrimSpace(row[5])
		}
		item := &model.FoodItem{
			FamilyID: familyID, Name: name, Category: category, Quantity: quantity,
			Unit: unit, ShelfLifeDays: days, StorageLocation: location, CreatorID: userID,
		}
		item.ExpiryDate = s.calculator.CalculateExpiryDate(nil, days, nil)
		item.Status = s.calculator.ComputeFreshness("", item.ExpiryDate)
		if err := s.repo.Create(item); err != nil {
			return 0, nil, util.LogError(s.log, ctx, constants.LOG_FOOD_IMPORTED, fmt.Errorf("import csv food: %w", err))
		}
		created = append(created, *item)
		count++
	}
	s.log.InfoContext(ctx, constants.LOG_FOOD_IMPORTED, "family_id", familyID, "imported", count)
	return count, created, nil
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
