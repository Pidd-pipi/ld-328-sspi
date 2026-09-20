package util

import (
	"context"
	"time"
)

// Job 定时任务回调。
type Job func(ctx context.Context)

// Scheduler 基于 time.Ticker 的轻量任务调度器。
type Scheduler struct {
	interval time.Duration
	stopCh   chan struct{}
}

// NewScheduler 构造调度器。
func NewScheduler(interval time.Duration) *Scheduler {
	return &Scheduler{interval: interval, stopCh: make(chan struct{})}
}

// Start 启动定时任务（goroutine 中执行，不阻塞调用方）。
func (s *Scheduler) Start(ctx context.Context, job Job) {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				job(ctx)
			case <-s.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop 停止调度器。
func (s *Scheduler) Stop() { close(s.stopCh) }
