package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
)

// ConsumptionRecordService 消耗记录服务：历史查询与频率分析。
type ConsumptionRecordService struct {
	repo      *repository.ConsumptionRecordRepository
	foodRepo  *repository.FoodItemRepository
	familySvc *FamilyGroupService
	log       *slog.Logger
}

// NewConsumptionRecordService 构造消耗记录服务。
func NewConsumptionRecordService(repo *repository.ConsumptionRecordRepository, foodRepo *repository.FoodItemRepository, familySvc *FamilyGroupService, log *slog.Logger) *ConsumptionRecordService {
	return &ConsumptionRecordService{repo: repo, foodRepo: foodRepo, familySvc: familySvc, log: log}
}

// List 分页查询家庭消耗记录。
func (s *ConsumptionRecordService) List(ctx context.Context, userID, familyID uint, page, pageSize int) ([]model.ConsumptionRecord, int64, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return nil, 0, err
	}
	records, total, err := s.repo.ListByFamily(familyID, page, pageSize)
	if err != nil {
		return nil, 0, util.LogError(s.log, ctx, constants.LOG_CONSUMPTION_ANALYSIS, fmt.Errorf("list consumption records: %w", err))
	}
	return records, total, nil
}

// ListByFood 查询某食品消耗历史。
func (s *ConsumptionRecordService) ListByFood(ctx context.Context, userID, foodID uint) ([]model.ConsumptionRecord, error) {
	item, err := s.foodRepo.FindByID(foodID)
	if err != nil {
		return nil, util.NotFoundError("食品（FoodItem）不存在", err)
	}
	if err := s.familySvc.IsMember(ctx, item.FamilyID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByFood(foodID)
}

// Analysis 消耗频率分析：按类别统计 + Top 排行。
func (s *ConsumptionRecordService) Analysis(ctx context.Context, userID, familyID uint, month string) (*ConsumptionAnalysis, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return nil, err
	}
	byCategory, err := s.repo.MonthlyStats(familyID, month)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_CONSUMPTION_ANALYSIS, fmt.Errorf("monthly stats: %w", err))
	}
	top, err := s.repo.TopConsumedFoods(familyID, month, 10)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_CONSUMPTION_ANALYSIS, fmt.Errorf("top foods: %w", err))
	}
	result := &ConsumptionAnalysis{Month: month, ByCategory: byCategory, TopConsumed: top}
	s.log.InfoContext(ctx, constants.LOG_CONSUMPTION_ANALYSIS, "family_id", familyID, "month", month)
	return result, nil
}

// ConsumptionAnalysis 消耗分析结果。
type ConsumptionAnalysis struct {
	Month       string             `json:"month"`
	ByCategory  []model.MonthlyConsumption `json:"by_category"`
	TopConsumed []model.TopFood    `json:"top_consumed"`
}

// Record 直接创建消耗记录（供 Consume 流程之外的补录）。
func (s *ConsumptionRecordService) Record(ctx context.Context, foodID, userID uint, quantity float64, consumedAt time.Time) (*model.ConsumptionRecord, error) {
	record := &model.ConsumptionRecord{FoodItemID: foodID, Quantity: quantity, UserID: userID, ConsumedAt: consumedAt}
	if err := s.repo.Create(record); err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_CONSUMPTION_RECORDED, fmt.Errorf("create consumption record: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_CONSUMPTION_RECORDED, "food_id", foodID, "quantity", quantity)
	return record, nil
}
