package util

import (
	"context"
	"log/slog"
	"os"
)

// NewLogger 构造 JSON 结构化日志（所有 handler/service/middleware 共享）。
func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

// LogError 记录错误日志并返回原错误（不吞错）。
func LogError(log *slog.Logger, ctx context.Context, template string, err error, attrs ...any) error {
	if err != nil {
		log.ErrorContext(ctx, template, append(attrs, "error", err.Error())...)
	}
	return err
}
