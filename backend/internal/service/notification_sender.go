package service

import (
	"context"
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/constants"
)

// NotificationSender 邮件模拟发送器：把站内通知以日志形式"发送"邮件。
type NotificationSender struct {
	log *slog.Logger
}

// NewNotificationSender 构造邮件模拟发送器。
func NewNotificationSender(log *slog.Logger) *NotificationSender {
	return &NotificationSender{log: log}
}

// SendEmail 模拟发送邮件（仅记录日志，接入真实邮件服务时替换实现）。
func (s *NotificationSender) SendEmail(ctx context.Context, to, subject, body string) {
	s.log.InfoContext(ctx, constants.LOG_EMAIL_SIMULATED,
		"to", to, "subject", subject, "body_len", len(body))
}
