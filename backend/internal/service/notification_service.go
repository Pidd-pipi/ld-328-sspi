package service

import (
	"time"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
)

// NotificationService 通知服务：查询、已读、临期/过期通知。
type NotificationService struct {
	repo      *repository.NotificationRepository
	familySvc *FamilyGroupService
	log       *slog.Logger
}

// NewNotificationService 构造通知服务。
func NewNotificationService(repo *repository.NotificationRepository, familySvc *FamilyGroupService, log *slog.Logger) *NotificationService {
	return &NotificationService{repo: repo, familySvc: familySvc, log: log}
}

// List 查询家庭通知。
func (s *NotificationService) List(ctx context.Context, userID, familyID uint, unreadOnly bool, page, pageSize int) ([]model.Notification, int64, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return nil, 0, err
	}
	items, total, err := s.repo.ListByFamily(familyID, unreadOnly, page, pageSize)
	if err != nil {
		return nil, 0, util.LogError(s.log, ctx, constants.LOG_NOTIFICATION_CREATED, fmt.Errorf("list notifications: %w", err))
	}
	return items, total, nil
}

// MarkRead 标记单条已读。
func (s *NotificationService) MarkRead(ctx context.Context, userID, id uint) error {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return util.NotFoundError("通知（Notification）不存在", err)
	}
	if err := s.familySvc.IsMember(ctx, n.FamilyID, userID); err != nil {
		return err
	}
	if err := s.repo.MarkRead(id); err != nil {
		return util.LogError(s.log, ctx, constants.LOG_NOTIFICATION_READ, fmt.Errorf("mark read: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_NOTIFICATION_READ, "notification_id", id)
	return nil
}

// MarkAllRead 标记家庭全部已读。
func (s *NotificationService) MarkAllRead(ctx context.Context, userID, familyID uint) error {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return err
	}
	if err := s.repo.MarkAllReadByFamily(familyID); err != nil {
		return util.LogError(s.log, ctx, constants.LOG_NOTIFICATION_READ, fmt.Errorf("mark all read: %w", err))
	}
	return nil
}

// UnreadCount 未读数量。
func (s *NotificationService) UnreadCount(ctx context.Context, userID, familyID uint) (int64, error) {
	if err := s.familySvc.IsMember(ctx, familyID, userID); err != nil {
		return 0, err
	}
	return s.repo.CountUnreadByFamily(familyID)
}

// Create 创建通知（供扫描器与邮件模拟使用）。
func (s *NotificationService) Create(ctx context.Context, familyID, foodItemID uint, typ, title, content string) error {
	if typ != constants.NotificationExpiring && typ != constants.NotificationExpired {
		return util.BadRequest("通知类型（Notification.type）不合法", errors.New("invalid type"))
	}
	exists, err := s.repo.HasForFoodAndType(foodItemID, typ)
	if err != nil {
		return err
	}
	if exists {
		return nil // 已存在同类型通知，避免重复
	}
	n := &model.Notification{
		FamilyID: familyID, FoodItemID: foodItemID, Type: typ,
		Title: title, Content: content, SendAt: now(),
	}
	if err := s.repo.Create(n); err != nil {
		return util.LogError(s.log, ctx, constants.LOG_NOTIFICATION_CREATED, fmt.Errorf("create notification: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_NOTIFICATION_CREATED, "food_item_id", foodItemID, "type", typ)
	return nil
}

func now() time.Time { return time.Now() }
