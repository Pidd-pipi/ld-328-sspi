package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
)

// StatsService 统计服务：看板与分类统计报表。
type StatsService struct {
	foodRepo        *repository.FoodItemRepository
	consumeRepo     *repository.ConsumptionRecordRepository
	notifyRepo      *repository.NotificationRepository
	familySvc       *FamilyGroupService
	memberSvc       *FamilyMemberService
	calculator      *util.FoodCalculator
	log             *slog.Logger
}

// NewStatsService 构造统计服务。
func NewStatsService(foodRepo *repository.FoodItemRepository, consumeRepo *repository.ConsumptionRecordRepository, notifyRepo *repository.NotificationRepository, familySvc *FamilyGroupService, memberSvc *FamilyMemberService, calculator *util.FoodCalculator, log *slog.Logger) *StatsService {
	return &StatsService{foodRepo: foodRepo, consumeRepo: consumeRepo, notifyRepo: notifyRepo, familySvc: familySvc, memberSvc: memberSvc, calculator: calculator, log: log}
}

// DashboardData 看板数据。
type DashboardData struct {
	TotalItems      int64              `json:"total_items"`
	ExpiringCount   int64              `json:"expiring_count"`
	ExpiredCount    int64              `json:"expired_count"`
	ConsumedCount   int64              `json:"consumed_count"`
	UnreadNotify    int64              `json:"unread_notify"`
	MemberCount     int64              `json:"member_count"`
	ByCategory      []model.CategoryCount `json:"by_category"`
	ExpiringItems   []model.FoodItem   `json:"expiring_items"`
	ExpiredItems    []model.FoodItem   `json:"expired_items"`
	RecentNotify    []model.Notification `json:"recent_notify"`
}

// Dashboard 生成看板数据。
func (s *StatsService) Dashboard(ctx context.Context, userID, familyID uint) (*DashboardData, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return nil, err
	}
	items, _, err := s.foodRepo.List(familyID, "", "", "", "", 1, 1000)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_STATS_DASHBOARD, fmt.Errorf("list foods: %w", err))
	}
	data := &DashboardData{}
	expiring := make([]model.FoodItem, 0)
	expired := make([]model.FoodItem, 0)
	consumed := int64(0)
	catMap := map[string]int64{}
	catQty := map[string]float64{}
	for i := range items {
		items[i].Status = s.calculator.ComputeFreshness(items[i].Status, items[i].ExpiryDate)
		switch items[i].Status {
		case constants.FreshnessExpiring:
			data.ExpiringCount++
			expiring = append(expiring, items[i])
		case constants.FreshnessExpired:
			data.ExpiredCount++
			expired = append(expired, items[i])
		case constants.FreshnessConsumed:
			consumed++
		}
		catMap[items[i].Category]++
		catQty[items[i].Category] += items[i].Quantity
	}
	data.TotalItems = int64(len(items))
	data.ConsumedCount = consumed
	data.ByCategory = make([]model.CategoryCount, 0, len(catMap))
	for cat, count := range catMap {
		data.ByCategory = append(data.ByCategory, model.CategoryCount{Category: cat, Count: count, TotalQuantity: catQty[cat]})
	}
	data.ExpiringItems = expiring
	data.ExpiredItems = expired
	notify, _, err := s.notifyRepo.ListByFamily(familyID, false, 1, 5)
	if err == nil {
		data.RecentNotify = notify
	}
	unread, _ := s.notifyRepo.CountUnreadByFamily(familyID)
	data.UnreadNotify = unread
	memberCount, _ := s.memberSvc.CountByFamily(ctx, familyID)
	data.MemberCount = memberCount
	s.log.InfoContext(ctx, constants.LOG_STATS_DASHBOARD, "family_id", familyID)
	return data, nil
}

// StatisticsData 分类统计报表。
type StatisticsData struct {
	CategoryShare    []model.CategoryCount `json:"category_share"`
	ConsumptionShare []model.MonthlyConsumption `json:"consumption_share"`
	TopPurchased     []model.TopFood       `json:"top_purchased"`
	TopWasted        []model.TopFood       `json:"top_wasted"`
	WasteAmount      float64               `json:"waste_amount"`
	Month            string                `json:"month"`
}

// Statistics 生成分类统计（按月份筛选）。
func (s *StatsService) Statistics(ctx context.Context, userID, familyID uint, month string) (*StatisticsData, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return nil, err
	}
	categoryShare, err := s.foodRepo.CountGroupByCategory(familyID)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_STATS_REPORT, fmt.Errorf("category share: %w", err))
	}
	consumptionShare, err := s.consumeRepo.MonthlyStats(familyID, month)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_STATS_REPORT, fmt.Errorf("consumption share: %w", err))
	}
	topPurchased, err := s.consumeRepo.TopConsumedFoods(familyID, month, 10)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_STATS_REPORT, fmt.Errorf("top consumed: %w", err))
	}
	allItems, _, err := s.foodRepo.List(familyID, "", "", "", "", 1, 1000)
	if err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_STATS_REPORT, fmt.Errorf("list all foods: %w", err))
	}
	// 浪费金额估算：实时计算新鲜度，仅统计已过期食品（按平均单价 15 元/单位估算）
	wasteAmount := 0.0
	topWasted := make([]model.TopFood, 0, 8)
	for _, it := range allItems {
		if s.calculator.ComputeFreshness(it.Status, it.ExpiryDate) == constants.FreshnessExpired {
			wasteAmount += it.Quantity * 15
			topWasted = append(topWasted, model.TopFood{FoodItemID: it.ID, Name: it.Name, Count: 1, Quantity: it.Quantity})
		}
	}
	result := &StatisticsData{
		CategoryShare: categoryShare, ConsumptionShare: consumptionShare,
		TopPurchased: topPurchased, TopWasted: topWasted, WasteAmount: wasteAmount, Month: month,
	}
	s.log.InfoContext(ctx, constants.LOG_STATS_REPORT, "family_id", familyID, "month", month)
	return result, nil
}
