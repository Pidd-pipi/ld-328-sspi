package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/redis/go-redis/v9"
)

// ReminderScheduler 临期扫描调度器：Redis 分布式锁 + time.Ticker。
type ReminderScheduler struct {
	reminderSvc *ReminderService
	scheduler   *util.Scheduler
	redis       *redis.Client
	log         *slog.Logger
}

// NewReminderScheduler 构造调度器。
func NewReminderScheduler(reminderSvc *ReminderService, redis *redis.Client, interval time.Duration, log *slog.Logger) *ReminderScheduler {
	return &ReminderScheduler{
		reminderSvc: reminderSvc,
		scheduler:   util.NewScheduler(interval),
		redis:       redis,
		log:         log,
	}
}

// Start 启动定时扫描。
func (s *ReminderScheduler) Start(ctx context.Context) {
	s.scheduler.Start(ctx, func(ctx context.Context) {
		// 尝试获取分布式锁，防止多实例重复扫描
		ok, err := s.redis.SetNX(ctx, constants.ReminderScanLockKey, "1", 30*time.Second).Result()
		if err != nil {
			s.log.WarnContext(ctx, constants.LOG_EXPIRY_SCAN_FAILED, "error", err)
			return
		}
		if !ok {
			return
		}
		defer s.redis.Del(ctx, constants.ReminderScanLockKey)
		if _, err := s.reminderSvc.Scan(ctx); err != nil {
			s.log.WarnContext(ctx, constants.LOG_EXPIRY_SCAN_FAILED, "error", err)
		}
	})
}

// Stop 停止调度。
func (s *ReminderScheduler) Stop() { s.scheduler.Stop() }
