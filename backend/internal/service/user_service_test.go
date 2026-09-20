package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// 命名内存库 + cache=shared：连接池内多个连接访问同一个内存库（隔离靠唯一名），
	// WAL/忙等待使并发事务测试不被 SQLITE_BUSY 干扰。
	dsn := fmt.Sprintf("file:cyfreshfood_test_%d?mode=memory&cache=shared&_busy_timeout=10000&_journal_mode=WAL&_foreign_keys=1", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.FamilyGroup{}, &model.FamilyMember{},
		&model.FoodItem{}, &model.ConsumptionRecord{}, &model.Notification{}, &model.Recipe{},
		&model.DisposalApplication{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_disposal_pending_food
		ON disposal_applications(food_item_id) WHERE status = 'pending'`).Error; err != nil {
		t.Fatalf("create disposal index: %v", err)
	}
	return db
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

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
