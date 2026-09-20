package service

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// disposalFixture 处置申请测试夹具。
type disposalFixture struct {
	db          *gorm.DB
	svc         *DisposalRequestService
	foodRepo    *repository.FoodItemRepository
	consumeRepo *repository.ConsumptionRecordRepository
	reqRepo     *repository.DisposalRequestRepository
	adminID     uint
	memberID    uint
	familyID    uint
}

func newDisposalFixture(t *testing.T) *disposalFixture {
	t.Helper()
	// 文件库 + WAL，真实模拟并发批准下的事务串行化
	dsn := filepath.Join(t.TempDir(), "disposal_test.db") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.FamilyGroup{}, &model.FamilyMember{},
		&model.FoodItem{}, &model.ConsumptionRecord{}, &model.DisposalRequest{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX uq_disposal_pending_food
		ON disposal_requests(food_item_id) WHERE status = 'pending'`).Error; err != nil {
		t.Fatalf("create partial index: %v", err)
	}

	admin := model.User{Phone: "13800000001", PasswordHash: "x", Name: "管理员", Role: constants.RoleAdmin}
	member := model.User{Phone: "13800000002", PasswordHash: "x", Name: "成员", Role: constants.RoleMember}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("create member: %v", err)
	}
	group := model.FamilyGroup{Name: "测试之家", OwnerID: admin.ID, InviteCode: "FAMTEST"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := db.Create(&[]model.FamilyMember{
		{FamilyID: group.ID, UserID: admin.ID, Role: constants.FamilyRoleAdmin},
		{FamilyID: group.ID, UserID: member.ID, Role: constants.FamilyRoleMember},
	}).Error; err != nil {
		t.Fatalf("create members: %v", err)
	}

	log := util.NewLogger()
	calc := util.NewFoodCalculator()
	foodRepo := repository.NewFoodItemRepository(db)
	consumeRepo := repository.NewConsumptionRecordRepository(db)
	reqRepo := repository.NewDisposalRequestRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	groupRepo := repository.NewFamilyGroupRepository(db)
	familySvc := NewFamilyGroupService(groupRepo, memberRepo, log)
	svc := NewDisposalRequestService(reqRepo, foodRepo, consumeRepo, familySvc, calc, log)

	return &disposalFixture{
		db: db, svc: svc, foodRepo: foodRepo, consumeRepo: consumeRepo, reqRepo: reqRepo,
		adminID: admin.ID, memberID: member.ID, familyID: group.ID,
	}
}

// seedFood 创建一条指定状态/余量/到期日的食品。
func (f *disposalFixture) seedFood(t *testing.T, name, status string, quantity float64, expiry time.Time) *model.FoodItem {
	t.Helper()
	item := &model.FoodItem{
		FamilyID: f.familyID, Name: name, Category: constants.FoodCategoryDairy,
		Quantity: quantity, Unit: "盒", StorageLocation: constants.StorageFridge,
		Status: status, CreatorID: f.memberID, ExpiryDate: &expiry,
	}
	if err := f.foodRepo.Create(item); err != nil {
		t.Fatalf("create food: %v", err)
	}
	return item
}

func TestDisposalService_CreateAndApproveFullDeduct(t *testing.T) {
	f := newDisposalFixture(t)
	ctx := context.Background()
	expired := time.Now().AddDate(0, 0, -1)
	item := f.seedFood(t, "过期酸奶", constants.FreshnessFresh, 2, expired)

	req, err := f.svc.Create(ctx, f.memberID, CreateDisposalInput{
		FoodItemID: item.ID, Quantity: 2, Method: constants.DisposalMethodDiscard, Reason: "已过期",
	})
	if err != nil {
		t.Fatalf("create disposal: %v", err)
	}
	if req.Status != constants.DisposalStatusPending {
		t.Fatalf("new request status = %s, want pending", req.Status)
	}

	approved, err := f.svc.Approve(ctx, f.adminID, req.ID)
	if err != nil {
		t.Fatalf("approve disposal: %v", err)
	}
	if approved.Status != constants.DisposalStatusApproved || approved.ConsumptionRecordID == nil {
		t.Fatalf("approved request not closed correctly: %+v", approved)
	}

	after, _ := f.foodRepo.FindByID(item.ID)
	if after.Quantity != 0 || after.Status != constants.FreshnessConsumed {
		t.Fatalf("food after approve = (%.2f,%s), want (0,consumed)", after.Quantity, after.Status)
	}
	var records []model.ConsumptionRecord
	if err := f.db.Where("food_item_id = ?", item.ID).Find(&records).Error; err != nil {
		t.Fatalf("query records: %v", err)
	}
	if len(records) != 1 || records[0].Quantity != 2 || records[0].UserID != f.adminID {
		t.Fatalf("consumption record not generated: %+v", records)
	}
}

func TestDisposalService_PartialApproveKeepsStatus(t *testing.T) {
	f := newDisposalFixture(t)
	ctx := context.Background()
	expiring := time.Now().AddDate(0, 0, 1)
	item := f.seedFood(t, "临期牛奶", constants.FreshnessFresh, 3, expiring)

	req, err := f.svc.Create(ctx, f.memberID, CreateDisposalInput{
		FoodItemID: item.ID, Quantity: 1, Method: constants.DisposalMethodDonate,
	})
	if err != nil {
		t.Fatalf("create disposal: %v", err)
	}
	if _, err := f.svc.Approve(ctx, f.adminID, req.ID); err != nil {
		t.Fatalf("approve: %v", err)
	}
	after, _ := f.foodRepo.FindByID(item.ID)
	if after.Quantity != 2 {
		t.Fatalf("quantity = %.2f, want 2", after.Quantity)
	}
	if after.Status == constants.FreshnessConsumed {
		t.Fatalf("partial approval must not mark food consumed, got %s", after.Status)
	}
}

func TestDisposalService_CreateRules(t *testing.T) {
	f := newDisposalFixture(t)
	ctx := context.Background()
	now := time.Now()
	expiring := now.AddDate(0, 0, 1)
	fresh := now.AddDate(0, 0, 10)

	expiredItem := f.seedFood(t, "过期品", constants.FreshnessFresh, 1, now.AddDate(0, 0, -1))
	freshItem := f.seedFood(t, "充裕品", constants.FreshnessFresh, 5, fresh)
	consumedItem := f.seedFood(t, "已消耗", constants.FreshnessConsumed, 0, expiring)
	overItem := f.seedFood(t, "超量品", constants.FreshnessFresh, 1, expiring)

	tests := []struct {
		name    string
		input   CreateDisposalInput
		wantErr bool
	}{
		{name: "fresh food not eligible", input: CreateDisposalInput{FoodItemID: freshItem.ID, Quantity: 1, Method: constants.DisposalMethodDiscard}, wantErr: true},
		{name: "consumed food rejected", input: CreateDisposalInput{FoodItemID: consumedItem.ID, Quantity: 1, Method: constants.DisposalMethodDiscard}, wantErr: true},
		{name: "quantity exceeds stock", input: CreateDisposalInput{FoodItemID: overItem.ID, Quantity: 2, Method: constants.DisposalMethodConsume}, wantErr: true},
		{name: "invalid method", input: CreateDisposalInput{FoodItemID: expiredItem.ID, Quantity: 1, Method: "burn"}, wantErr: true},
		{name: "zero quantity", input: CreateDisposalInput{FoodItemID: expiredItem.ID, Quantity: 0, Method: "discard"}, wantErr: true},
		{name: "valid expiring food", input: CreateDisposalInput{FoodItemID: overItem.ID, Quantity: 1, Method: "consume"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.svc.Create(ctx, f.memberID, tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDisposalService_DuplicatePendingRejected(t *testing.T) {
	f := newDisposalFixture(t)
	ctx := context.Background()
	item := f.seedFood(t, "唯一待处理", constants.FreshnessFresh, 3, time.Now().AddDate(0, 0, -2))

	if _, err := f.svc.Create(ctx, f.memberID, CreateDisposalInput{FoodItemID: item.ID, Quantity: 1, Method: "discard"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	// 同一食品第二张待处理申请：应用层 + 部分唯一索引双重拒绝
	if _, err := f.svc.Create(ctx, f.memberID, CreateDisposalInput{FoodItemID: item.ID, Quantity: 1, Method: "donate"}); err == nil {
		t.Fatalf("duplicate pending request must be rejected")
	}

	// 结案后允许重新申请
	reqs, _, err := f.reqRepo.List(f.familyID, item.ID, constants.DisposalStatusPending, 1, 10)
	if err != nil || len(reqs) != 1 {
		t.Fatalf("pending list = %d, err=%v", len(reqs), err)
	}
	if _, err := f.svc.Reject(ctx, f.adminID, reqs[0].ID, "不需要处理"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	if _, err := f.svc.Create(ctx, f.memberID, CreateDisposalInput{FoodItemID: item.ID, Quantity: 1, Method: "donate"}); err != nil {
		t.Fatalf("create after rejection should succeed: %v", err)
	}
}

func TestDisposalService_RejectDoesNotTouchStock(t *testing.T) {
	f := newDisposalFixture(t)
	ctx := context.Background()
	item := f.seedFood(t, "驳回品", constants.FreshnessFresh, 2, time.Now().AddDate(0, 0, -1))
	req, err := f.svc.Create(ctx, f.memberID, CreateDisposalInput{FoodItemID: item.ID, Quantity: 2, Method: "discard"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := f.svc.Reject(ctx, f.adminID, req.ID, "继续食用"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	after, _ := f.foodRepo.FindByID(item.ID)
	if after.Quantity != 2 || after.Status == constants.FreshnessConsumed {
		t.Fatalf("reject changed stock: %.2f %s", after.Quantity, after.Status)
	}
	var cnt int64
	f.db.Model(&model.ConsumptionRecord{}).Where("food_item_id = ?", item.ID).Count(&cnt)
	if cnt != 0 {
		t.Fatalf("reject generated consumption records: %d", cnt)
	}
}

func TestDisposalService_MemberCannotReview(t *testing.T) {
	f := newDisposalFixture(t)
	ctx := context.Background()
	item := f.seedFood(t, "权限品", constants.FreshnessFresh, 1, time.Now().AddDate(0, 0, -1))
	req, err := f.svc.Create(ctx, f.memberID, CreateDisposalInput{FoodItemID: item.ID, Quantity: 1, Method: "discard"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := f.svc.Approve(ctx, f.memberID, req.ID); err == nil {
		t.Fatalf("member must not approve")
	}
	if _, err := f.svc.Reject(ctx, f.memberID, req.ID, "x"); err == nil {
		t.Fatalf("member must not reject")
	}
	// 被拒后申请仍为待处理
	pending, _ := f.reqRepo.FindByID(req.ID)
	if pending.Status != constants.DisposalStatusPending {
		t.Fatalf("status = %s, want pending", pending.Status)
	}
}

func TestDisposalService_ConcurrentApproveOnlyOnce(t *testing.T) {
	f := newDisposalFixture(t)
	ctx := context.Background()
	item := f.seedFood(t, "并发品", constants.FreshnessFresh, 5, time.Now().AddDate(0, 0, -1))
	req, err := f.svc.Create(ctx, f.memberID, CreateDisposalInput{FoodItemID: item.ID, Quantity: 3, Method: "discard"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	const n = 8
	var wg sync.WaitGroup
	start := make(chan struct{})
	var okCount, failCount int64
	var mu sync.Mutex
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			_, err := f.svc.Approve(ctx, f.adminID, req.ID)
			mu.Lock()
			if err == nil {
				okCount++
			} else {
				failCount++
			}
			mu.Unlock()
		}()
	}
	close(start)
	wg.Wait()

	if okCount != 1 {
		t.Fatalf("successful approvals = %d, want exactly 1 (failures=%d)", okCount, failCount)
	}
	after, _ := f.foodRepo.FindByID(item.ID)
	if after.Quantity != 2 {
		t.Fatalf("quantity = %.2f, want 2 (failed approvals must not change stock)", after.Quantity)
	}
	var recordCnt int64
	f.db.Model(&model.ConsumptionRecord{}).Where("food_item_id = ?", item.ID).Count(&recordCnt)
	if recordCnt != 1 {
		t.Fatalf("consumption records = %d, want exactly 1", recordCnt)
	}
	closed, _ := f.reqRepo.FindByID(req.ID)
	if closed.Status != constants.DisposalStatusApproved {
		t.Fatalf("final status = %s, want approved", closed.Status)
	}
}

func TestDisposalService_ApproveConsumedFoodRejectsWholeOrder(t *testing.T) {
	f := newDisposalFixture(t)
	ctx := context.Background()
	item := f.seedFood(t, "审批时已消耗", constants.FreshnessFresh, 2, time.Now().AddDate(0, 0, -1))
	req, err := f.svc.Create(ctx, f.memberID, CreateDisposalInput{FoodItemID: item.ID, Quantity: 2, Method: "consume"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// 其他流程先行消耗（食品余量归零）
	if _, err := f.foodRepo.FindByID(item.ID); err != nil {
		t.Fatalf("find: %v", err)
	}
	if err := f.db.Model(&model.FoodItem{}).Where("id = ?", item.ID).
		Updates(map[string]interface{}{"quantity": 0, "status": constants.FreshnessConsumed}).Error; err != nil {
		t.Fatalf("consume food: %v", err)
	}
	if _, err := f.svc.Approve(ctx, f.adminID, req.ID); err == nil {
		t.Fatalf("approving against consumed food must reject the whole order")
	}
	closed, _ := f.reqRepo.FindByID(req.ID)
	if closed.Status != constants.DisposalStatusPending {
		t.Fatalf("failed approval must not close the request, got %s", closed.Status)
	}
}
