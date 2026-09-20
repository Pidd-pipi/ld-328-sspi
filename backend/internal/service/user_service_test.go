package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.FamilyGroup{}, &model.FamilyMember{},
		&model.FoodItem{}, &model.ConsumptionRecord{}, &model.Notification{}, &model.Recipe{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func testLogger() *slog.Logger { return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})) }

func TestUserService_RegisterAndLogin(t *testing.T) {
	db := newTestDB(t)
	svc := NewUserService(repository.NewUserRepository(db), "test-secret", 24, testLogger())
	ctx := context.Background()

	user, token, err := svc.Register(ctx, "13900000001", "pass123", "测试用户")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if token == "" || user.Phone != "13900000001" {
		t.Fatalf("register result mismatch: user=%+v token=%q", user, token)
	}

	// 重复手机号应返回冲突
	if _, _, err := svc.Register(ctx, "13900000001", "pass456", "重复用户"); err == nil {
		t.Fatal("expected conflict error for duplicate phone")
	}

	// 错误密码
	if _, _, err := svc.Login(ctx, "13900000001", "wrong"); err == nil {
		t.Fatal("expected login failure for wrong password")
	}

	// 正确登录
	got, token2, err := svc.Login(ctx, "13900000001", "pass123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if got.ID != user.ID || token2 == "" {
		t.Fatalf("login result mismatch")
	}
}
