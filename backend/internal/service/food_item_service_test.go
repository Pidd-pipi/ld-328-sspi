package service

import (
	"context"
	"testing"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

func seedFood(t *testing.T, db *gorm.DB, familyID, creatorID uint, quantity float64) *model.FoodItem {
	t.Helper()
	expiry := time.Now().AddDate(0, 0, 5)
	item := &model.FoodItem{
		FamilyID: familyID, Name: "测试食品", Category: constants.FoodCategoryDairy,
		Quantity: quantity, Unit: "盒", ShelfLifeDays: 5, StorageLocation: constants.StorageFridge,
		Status: constants.FreshnessFresh, CreatorID: creatorID, ExpiryDate: &expiry,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("seed food: %v", err)
	}
	return item
}

func TestFoodItemService_Consume(t *testing.T) {
	db := newTestDB(t)
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	foodRepo := repository.NewFoodItemRepository(db)
	consumeRepo := repository.NewConsumptionRecordRepository(db)
	familySvc := NewFamilyGroupService(groupRepo, memberRepo, testLogger())
	svc := NewFoodItemService(foodRepo, consumeRepo, familySvc, util.NewFoodCalculator(), testLogger())
	ctx := context.Background()

	group, err := familySvc.Create(ctx, 1, "测试家庭")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	item := seedFood(t, db, group.ID, 1, 3)

	record, err := svc.Consume(ctx, 1, item.ID, 1.5, time.Now())
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if record.Quantity != 1.5 {
		t.Fatalf("record quantity = %v, want 1.5", record.Quantity)
	}

	got, err := foodRepo.FindByID(item.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Quantity != 1.5 {
		t.Fatalf("remaining quantity = %v, want 1.5", got.Quantity)
	}

	// 全部消耗后状态流转为 consumed。
	if _, err := svc.Consume(ctx, 1, item.ID, 1.5, time.Now()); err != nil {
		t.Fatalf("second Consume() error = %v", err)
	}
	got, _ = foodRepo.FindByID(item.ID)
	if got.Status != constants.FreshnessConsumed {
		t.Fatalf("status = %s, want %s", got.Status, constants.FreshnessConsumed)
	}

	// 超量消耗应失败。
	if _, err := svc.Consume(ctx, 1, item.ID, 0.1, time.Now()); err == nil {
		t.Fatal("expected error for consuming consumed/empty food")
	}
}
