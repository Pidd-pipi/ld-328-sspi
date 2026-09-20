package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
	"gorm.io/gorm"
)

func seedDisposalFood(t *testing.T, db *gorm.DB, familyID, creatorID uint, quantity float64, daysToExpiry int, status string) *model.FoodItem {
	t.Helper()
	expiry := time.Now().AddDate(0, 0, daysToExpiry)
	item := &model.FoodItem{
		FamilyID: familyID, Name: "临期食品", Category: constants.FoodCategoryDairy,
		Quantity: quantity, Unit: "盒", ShelfLifeDays: 5, StorageLocation: constants.StorageFridge,
		Status: status, CreatorID: creatorID, ExpiryDate: &expiry,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("seed disposal food: %v", err)
	}
	return item
}

func newDisposalService(t *testing.T, db *gorm.DB) (*DisposalApplicationService, *FamilyGroupService, *repository.FoodItemRepository, *repository.ConsumptionRecordRepository) {
	t.Helper()
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	foodRepo := repository.NewFoodItemRepository(db)
	consumeRepo := repository.NewConsumptionRecordRepository(db)
	disposalRepo := repository.NewDisposalApplicationRepository(db)
	familySvc := NewFamilyGroupService(groupRepo, memberRepo, testLogger())
	svc := NewDisposalApplicationService(disposalRepo, foodRepo, consumeRepo, familySvc, util.NewFoodCalculator(), testLogger())
	return svc, familySvc, foodRepo, consumeRepo
}

func TestDisposalApplication_FullLifecycle(t *testing.T) {
	db := newTestDB(t)
	svc, familySvc, foodRepo, consumeRepo := newDisposalService(t, db)
	ctx := context.Background()

	group, err := familySvc.Create(ctx, 1, "处置测试家庭")
	if err != nil {
		t.Fatalf("Create group: %v", err)
	}
	// 用户 2 作为普通成员加入
	member2 := model.FamilyMember{FamilyID: group.ID, UserID: 2, Role: constants.FamilyRoleMember}
	if err := db.Create(&member2).Error; err != nil {
		t.Fatalf("add member: %v", err)
	}
	// 2 盒、1 天后到期 → 临期
	item := seedDisposalFood(t, db, group.ID, 1, 2, 1, constants.FreshnessFresh)

	// 成员提交申请
	app, err := svc.Create(ctx, 2, CreateDisposalInput{FoodItemID: item.ID, Quantity: 1, Method: constants.DisposalMethodDonate})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if app.Status != constants.DisposalStatusPending {
		t.Fatalf("status = %s, want pending", app.Status)
	}

	// 同一食品第二张待处理申请被拒绝（409）
	if _, err := svc.Create(ctx, 2, CreateDisposalInput{FoodItemID: item.ID, Quantity: 1, Method: constants.DisposalMethodEat}); err == nil {
		t.Fatal("expected duplicate pending application conflict")
	}

	// 普通成员不能批准（403）
	if _, err := svc.Approve(ctx, 2, app.ID, ""); err == nil {
		t.Fatal("member must not approve")
	}

	// 管理员批准
	approved, err := svc.Approve(ctx, 1, app.ID, "同意捐赠")
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if approved.Status != constants.DisposalStatusApproved {
		t.Fatalf("status = %s, want approved", approved.Status)
	}
	if approved.ReviewerID == nil || *approved.ReviewerID != 1 {
		t.Fatalf("reviewer = %v, want 1", approved.ReviewerID)
	}

	// 余量从 2 扣到 1，状态不变（未归零）
	got, _ := foodRepo.FindByID(item.ID)
	if got.Quantity != 1 {
		t.Fatalf("quantity = %v, want 1", got.Quantity)
	}
	if got.Status == constants.FreshnessConsumed {
		t.Fatal("food with remaining stock must not be marked consumed")
	}
	// 生成了一条消耗记录
	records, err := consumeRepo.ListByFood(item.ID)
	if err != nil || len(records) != 1 {
		t.Fatalf("consumption records = %d, err = %v; want 1", len(records), err)
	}

	// 并发重复批准必须失败且库存不变
	before := got.Quantity
	if _, err := svc.Approve(ctx, 1, app.ID, ""); err == nil {
		t.Fatal("second approve must fail")
	}
	again, _ := foodRepo.FindByID(item.ID)
	if again.Quantity != before {
		t.Fatalf("failed approve changed stock: %v -> %v", before, again.Quantity)
	}
}

func TestDisposalApplication_ApproveZeroMarksConsumed(t *testing.T) {
	db := newTestDB(t)
	svc, familySvc, foodRepo, _ := newDisposalService(t, db)
	ctx := context.Background()
	group, _ := familySvc.Create(ctx, 1, "归零家庭")
	item := seedDisposalFood(t, db, group.ID, 1, 1, 0, constants.FreshnessExpiring)

	app, err := svc.Create(ctx, 1, CreateDisposalInput{FoodItemID: item.ID, Quantity: 1, Method: constants.DisposalMethodDiscard})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Approve(ctx, 1, app.ID, ""); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	got, _ := foodRepo.FindByID(item.ID)
	if got.Quantity != 0 || got.Status != constants.FreshnessConsumed {
		t.Fatalf("after full approval quantity=%v status=%s, want 0/consumed", got.Quantity, got.Status)
	}
}

func TestDisposalApplication_OverQuantityRejectedWholeOrder(t *testing.T) {
	db := newTestDB(t)
	svc, familySvc, foodRepo, consumeRepo := newDisposalService(t, db)
	ctx := context.Background()
	group, _ := familySvc.Create(ctx, 1, "超量家庭")
	item := seedDisposalFood(t, db, group.ID, 1, 2, -1, constants.FreshnessExpired)

	// 申请 2 盒，批准前食品被另行消耗掉 1.5 盒 → 余量 0.5，批准整单拒绝
	app, err := svc.Create(ctx, 1, CreateDisposalInput{FoodItemID: item.ID, Quantity: 2, Method: constants.DisposalMethodEat})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := foodRepo.DeductQuantity(item.ID, 1.5); err != nil {
		t.Fatalf("pre-consume: %v", err)
	}
	if _, err := svc.Approve(ctx, 1, app.ID, ""); err == nil {
		t.Fatal("approve exceeding stock must fail")
	}
	// 失败不改库存，仍是 0.5；申请仍为待处理
	got, _ := foodRepo.FindByID(item.ID)
	if got.Quantity != 0.5 {
		t.Fatalf("stock = %v, want 0.5 unchanged", got.Quantity)
	}
	pending, _ := svc.GetByID(ctx, 1, app.ID)
	if pending.Status != constants.DisposalStatusPending {
		t.Fatalf("application status = %s, want pending", pending.Status)
	}
	recs, _ := consumeRepo.ListByFood(item.ID)
	if len(recs) != 0 { // 驳回的整单未写入任何消耗记录
		t.Fatalf("consumption records = %d, want 0 after rejected approval", len(recs))
	}
}

func TestDisposalApplication_RejectOnlyCloses(t *testing.T) {
	db := newTestDB(t)
	svc, familySvc, foodRepo, consumeRepo := newDisposalService(t, db)
	ctx := context.Background()
	group, _ := familySvc.Create(ctx, 1, "驳回家庭")
	item := seedDisposalFood(t, db, group.ID, 1, 3, 2, constants.FreshnessExpiring)

	app, err := svc.Create(ctx, 1, CreateDisposalInput{FoodItemID: item.ID, Quantity: 2, Method: constants.DisposalMethodDonate})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	rejected, err := svc.Reject(ctx, 1, app.ID, "自行处理")
	if err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if rejected.Status != constants.DisposalStatusRejected {
		t.Fatalf("status = %s, want rejected", rejected.Status)
	}
	got, _ := foodRepo.FindByID(item.ID)
	if got.Quantity != 3 {
		t.Fatalf("stock = %v, want 3 unchanged after reject", got.Quantity)
	}
	recs, _ := consumeRepo.ListByFood(item.ID)
	if len(recs) != 0 {
		t.Fatalf("consumption records = %d, want 0 after reject", len(recs))
	}
	// 驳回后可重新申请
	if _, err := svc.Create(ctx, 1, CreateDisposalInput{FoodItemID: item.ID, Quantity: 1, Method: constants.DisposalMethodEat}); err != nil {
		t.Fatalf("re-apply after reject error = %v", err)
	}
	// 再次驳回已结案申请应失败
	if _, err := svc.Reject(ctx, 1, app.ID, ""); err == nil {
		t.Fatal("rejecting closed application must fail")
	}
}

func TestDisposalApplication_FreshFoodNotAllowed(t *testing.T) {
	db := newTestDB(t)
	svc, familySvc, _, _ := newDisposalService(t, db)
	ctx := context.Background()
	group, _ := familySvc.Create(ctx, 1, "新鲜家庭")
	// 10 天后到期 → fresh，不可申请
	item := seedDisposalFood(t, db, group.ID, 1, 2, 10, constants.FreshnessFresh)

	_, err := svc.Create(ctx, 1, CreateDisposalInput{FoodItemID: item.ID, Quantity: 1, Method: constants.DisposalMethodDiscard})
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.Code != constants.CodeDisposalNotAllowed {
		t.Fatalf("err = %v, want CodeDisposalNotAllowed", err)
	}
}

func TestDisposalApplication_ConcurrentApproveOnlyOneWins(t *testing.T) {
	db := newTestDB(t)
	svc, familySvc, foodRepo, consumeRepo := newDisposalService(t, db)
	ctx := context.Background()
	group, _ := familySvc.Create(ctx, 1, "并发家庭")

	const n = 20
	for i := 0; i < 3; i++ {
		item := seedDisposalFood(t, db, group.ID, 1, 10, i-2, constants.FreshnessExpiring)
		app, err := svc.Create(ctx, 1, CreateDisposalInput{FoodItemID: item.ID, Quantity: 3, Method: constants.DisposalMethodEat})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		var wg sync.WaitGroup
		var successCount, stockMutations int64
		var mu sync.Mutex
		start := make(chan struct{})
		var firstErr error
		for j := 0; j < n; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				if _, err := svc.Approve(ctx, 1, app.ID, ""); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}
				mu.Lock()
				successCount++
				mu.Unlock()
			}()
		}
		close(start)
		wg.Wait()
		if successCount != 1 {
			t.Errorf("case %d: successCount = %d, want exactly 1; firstErr = %v", i, successCount, firstErr)
		}

		got, err := foodRepo.FindByID(item.ID)
		if err != nil {
			t.Fatalf("FindByID after concurrent approve: %v", err)
		}
		recs, _ := consumeRepo.ListByFood(item.ID)
		// 只统计本次批准生成的 3.0 数量记录，避免与其它用例交叉（每个食品独立）
		for _, r := range recs {
			if r.Quantity == 3 {
				stockMutations++
			}
		}

		if successCount != 1 {
			t.Errorf("case %d: successCount = %d, want exactly 1", i, successCount)
		}
		if got.Quantity != 7 {
			t.Errorf("case %d: quantity = %v, want 7 (deducted exactly once)", i, got.Quantity)
		}
		if stockMutations != 1 {
			t.Errorf("case %d: consumption records = %d, want exactly 1", i, stockMutations)
		}
	}
}
