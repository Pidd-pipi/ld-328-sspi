package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

// ReminderService 临期/过期扫描：刷新食品状态并生成站内通知。
type ReminderService struct {
	foodRepo   *repository.FoodItemRepository
	notifyRepo *repository.NotificationRepository
	calculator *util.FoodCalculator
	log        *slog.Logger
}

// NewReminderService 构造临期扫描服务。
func NewReminderService(foodRepo *repository.FoodItemRepository, notifyRepo *repository.NotificationRepository, calculator *util.FoodCalculator, log *slog.Logger) *ReminderService {
	return &ReminderService{foodRepo: foodRepo, notifyRepo: notifyRepo, calculator: calculator, log: log}
}

// Scan 执行一轮临期/过期扫描（全家庭），返回新增通知数。
func (s *ReminderService) Scan(ctx context.Context) (int, error) {
	s.log.InfoContext(ctx, constants.LOG_EXPIRY_SCAN_STARTED)
	items, err := s.foodRepo.ListByStatus(0, []string{constants.FreshnessFresh, constants.FreshnessExpiring})
	if err != nil {
		return 0, util.LogError(s.log, ctx, constants.LOG_EXPIRY_SCAN_FAILED, fmt.Errorf("scan foods: %w", err))
	}
	created := 0
	for _, item := range items {
		status := s.calculator.ComputeFreshness(item.Status, item.ExpiryDate)
		if status == item.Status {
			continue
		}
		typ := constants.NotificationExpiring
		title := constants.MsgFoodExpiring
		if status == constants.FreshnessExpired {
			typ = constants.NotificationExpired
			title = constants.MsgFoodExpired
		}
		content := fmt.Sprintf("%s 已%s，请及时处理。", item.Name, util.FreshnessStatusText(status))
		exists, _ := s.notifyRepo.HasForFoodAndType(item.ID, typ)
		if exists {
			continue
		}
		n := &model.Notification{
			FamilyID: item.FamilyID, FoodItemID: item.ID, Type: typ,
			Title: title, Content: content,
		}
		err = s.notifyRepo.Transaction(func(tx *gorm.DB) error {
			if err := s.foodRepo.WithTx(tx).UpdateStatus(item.ID, status); err != nil {
				return fmt.Errorf("update food status: %w", err)
			}
			if err := s.notifyRepo.WithTx(tx).Create(n); err != nil {
				return fmt.Errorf("create notification: %w", err)
			}
			return nil
		})
		if err != nil {
			return created, util.LogError(s.log, ctx, constants.LOG_EXPIRY_SCAN_FAILED, err)
		}
		s.log.InfoContext(ctx, constants.LOG_FOOD_STATUS_REFRESHED, "food_id", item.ID, "status", status)
		s.log.InfoContext(ctx, constants.LOG_NOTIFICATION_CREATED, "food_item_id", item.ID, "type", typ)
		created++
	}
	s.log.InfoContext(ctx, constants.LOG_EXPIRY_SCAN_FINISHED, "created", created)
	return created, nil
}
