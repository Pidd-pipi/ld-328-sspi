package router_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/blueship581/cyfreshfood/internal/config"
	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/handler"
	"github.com/blueship581/cyfreshfood/internal/middleware"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/router"
	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const testJWTSecret = "integration-test-secret"

type disposalAPICtx struct {
	t         *testing.T
	srv       http.Handler
	db        *gorm.DB
	adminTok  string
	memberTok string
	adminID   uint
	memberID  uint
	familyID  uint
}

type apiResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"message"`
	Data json.RawMessage `json:"data"`
}

func setupDisposalAPI(t *testing.T) *disposalAPICtx {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := filepath.Join(t.TempDir(), "api_test.db") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.FamilyGroup{}, &model.FamilyMember{}, &model.FoodItem{},
		&model.ConsumptionRecord{}, &model.Notification{}, &model.Recipe{}, &model.DisposalRequest{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX uq_disposal_pending_food
		ON disposal_requests(food_item_id) WHERE status = 'pending'`).Error; err != nil {
		t.Fatalf("index: %v", err)
	}

	admin := model.User{Phone: "13900000001", PasswordHash: "x", Name: "管理员", Role: constants.RoleAdmin}
	member := model.User{Phone: "13900000002", PasswordHash: "x", Name: "成员", Role: constants.RoleMember}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatal(err)
	}
	group := model.FamilyGroup{Name: "API 之家", OwnerID: admin.ID, InviteCode: "FAMAPI"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&[]model.FamilyMember{
		{FamilyID: group.ID, UserID: admin.ID, Role: constants.FamilyRoleAdmin},
		{FamilyID: group.ID, UserID: member.ID, Role: constants.FamilyRoleMember},
	}).Error; err != nil {
		t.Fatal(err)
	}

	log := util.NewLogger()
	calc := util.NewFoodCalculator()
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	foodRepo := repository.NewFoodItemRepository(db)
	consumeRepo := repository.NewConsumptionRecordRepository(db)
	disposalRepo := repository.NewDisposalRequestRepository(db)
	notifyRepo := repository.NewNotificationRepository(db)
	recipeRepo := repository.NewRecipeRepository(db)

	userSvc := service.NewUserService(userRepo, testJWTSecret, 72, log)
	familySvc := service.NewFamilyGroupService(groupRepo, memberRepo, log)
	memberSvc := service.NewFamilyMemberService(memberRepo, log)
	foodSvc := service.NewFoodItemService(foodRepo, consumeRepo, familySvc, calc, log)
	consumeSvc := service.NewConsumptionRecordService(consumeRepo, foodRepo, familySvc, log)
	disposalSvc := service.NewDisposalRequestService(disposalRepo, foodRepo, consumeRepo, familySvc, calc, log)
	notifySvc := service.NewNotificationService(notifyRepo, familySvc, log)
	recipeSvc := service.NewRecipeService(recipeRepo, foodRepo, familySvc, calc, log)
	statsSvc := service.NewStatsService(foodRepo, consumeRepo, notifyRepo, familySvc, memberSvc, calc, log)

	cfg := config.Config{JWTSecret: testJWTSecret, RateLimitPerMin: 100000}
	handlers := router.Handlers{
		User:         handler.NewUserHandler(userSvc, log),
		FamilyGroup:  handler.NewFamilyGroupHandler(familySvc, memberSvc, log),
		FoodItem:     handler.NewFoodItemHandler(foodSvc, consumeSvc, log),
		Consumption:  handler.NewConsumptionRecordHandler(consumeSvc, log),
		Disposal:     handler.NewDisposalRequestHandler(disposalSvc, log),
		Notification: handler.NewNotificationHandler(notifySvc, log),
		Recipe:       handler.NewRecipeHandler(recipeSvc, log),
		Stats:        handler.NewStatsHandler(statsSvc, log),
	}
	srv := router.New(cfg, log, handlers, middleware.NewRateLimiter(cfg.RateLimitPerMin))

	adminTok, err := util.GenerateToken(testJWTSecret, 72, admin.ID, admin.Phone, admin.Role)
	if err != nil {
		t.Fatal(err)
	}
	memberTok, err := util.GenerateToken(testJWTSecret, 72, member.ID, member.Phone, member.Role)
	if err != nil {
		t.Fatal(err)
	}
	return &disposalAPICtx{
		t: t, srv: srv, db: db,
		adminTok: adminTok, memberTok: memberTok,
		adminID: admin.ID, memberID: member.ID, familyID: group.ID,
	}
}

func (a *disposalAPICtx) seedFood(name, status string, qty float64, expiry time.Time) model.FoodItem {
	item := model.FoodItem{
		FamilyID: a.familyID, Name: name, Category: constants.FoodCategoryFresh,
		Quantity: qty, Unit: "份", StorageLocation: constants.StorageFridge,
		Status: status, CreatorID: a.memberID, ExpiryDate: &expiry,
	}
	if err := a.db.Create(&item).Error; err != nil {
		a.t.Fatalf("seed food: %v", err)
	}
	return item
}

func (a *disposalAPICtx) do(method, path, token string, body any) (int, apiResp) {
	var reader *bytes.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	a.srv.ServeHTTP(rec, req)
	var resp apiResp
	if rec.Body.Len() > 0 {
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	}
	return rec.Code, resp
}

func TestDisposalAPI_ApproveClosedLoop(t *testing.T) {
	a := setupDisposalAPI(t)
	expired := a.seedFood("过期熟食", constants.FreshnessFresh, 2, time.Now().AddDate(0, 0, -1))

	// 1. 成员提交申请
	status, resp := a.do(http.MethodPost, "/api/v1/disposals", a.memberTok, map[string]any{
		"food_item_id": expired.ID, "quantity": 2, "method": "discard", "reason": "已过期",
	})
	if status != http.StatusCreated || resp.Code != 0 {
		t.Fatalf("create: status=%d code=%d msg=%s", status, resp.Code, resp.Msg)
	}
	var req model.DisposalRequest
	_ = json.Unmarshal(resp.Data, &req)
	if req.Status != constants.DisposalStatusPending || req.ApplicantID != a.memberID {
		t.Fatalf("unexpected request: %+v", req)
	}

	// 2. 同一食品重复申请 → 409 / 1111
	if status, resp := a.do(http.MethodPost, "/api/v1/disposals", a.memberTok, map[string]any{
		"food_item_id": expired.ID, "quantity": 1, "method": "donate",
	}); status != http.StatusConflict || resp.Code != constants.CodeDisposalDuplicate {
		t.Fatalf("duplicate: status=%d code=%d", status, resp.Code)
	}

	// 3. 成员无权批准 → 403
	if status, _ := a.do(http.MethodPost, fmt.Sprintf("/api/v1/disposals/%d/approve", req.ID), a.memberTok, map[string]any{}); status != http.StatusForbidden {
		t.Fatalf("member approve status=%d, want 403", status)
	}

	// 4. 列表回读（待处理）
	status, resp = a.do(http.MethodGet,
		fmt.Sprintf("/api/v1/disposals?family_id=%d&status=pending", a.familyID), a.memberTok, nil)
	if status != http.StatusOK {
		t.Fatalf("list status=%d", status)
	}
	var page struct {
		List  []model.DisposalRequest `json:"list"`
		Total int64                   `json:"total"`
	}
	_ = json.Unmarshal(resp.Data, &page)
	if page.Total != 1 || len(page.List) != 1 || page.List[0].FoodItem.Name != "过期熟食" {
		t.Fatalf("unexpected list: %+v", page)
	}

	// 5. 管理员批准 → 扣减余量 + 消耗记录 + 结案
	status, resp = a.do(http.MethodPost, fmt.Sprintf("/api/v1/disposals/%d/approve", req.ID), a.adminTok, map[string]any{})
	if status != http.StatusOK || resp.Code != 0 {
		t.Fatalf("approve: status=%d code=%d msg=%s", status, resp.Code, resp.Msg)
	}
	var approved model.DisposalRequest
	_ = json.Unmarshal(resp.Data, &approved)
	if approved.Status != constants.DisposalStatusApproved || approved.ConsumptionRecordID == nil || approved.ReviewerID == nil || *approved.ReviewerID != a.adminID {
		t.Fatalf("approved request wrong: %+v", approved)
	}

	// 6. 并发/重复批准 → 409 / 1112，库存不再变化
	if status, resp := a.do(http.MethodPost, fmt.Sprintf("/api/v1/disposals/%d/approve", req.ID), a.adminTok, map[string]any{}); status != http.StatusConflict || resp.Code != constants.CodeDisposalNotPending {
		t.Fatalf("re-approve: status=%d code=%d", status, resp.Code)
	}

	// 7. 食品余量归零并标记已消耗
	status, resp = a.do(http.MethodGet, fmt.Sprintf("/api/v1/foods/%d", expired.ID), a.memberTok, nil)
	if status != http.StatusOK {
		t.Fatalf("food detail status=%d", status)
	}
	var detail struct {
		Item model.FoodItem `json:"item"`
	}
	_ = json.Unmarshal(resp.Data, &detail)
	if detail.Item.Quantity != 0 || detail.Item.Status != constants.FreshnessConsumed {
		t.Fatalf("food after approve = (%.2f,%s)", detail.Item.Quantity, detail.Item.Status)
	}

	// 8. 消耗记录可查询，数量与操作人正确
	status, resp = a.do(http.MethodGet, fmt.Sprintf("/api/v1/consumptions?family_id=%d", a.familyID), a.memberTok, nil)
	if status != http.StatusOK {
		t.Fatalf("consumptions status=%d", status)
	}
	var records struct {
		List []model.ConsumptionRecord `json:"list"`
	}
	_ = json.Unmarshal(resp.Data, &records)
	if len(records.List) != 1 || records.List[0].Quantity != 2 || records.List[0].UserID != a.adminID {
		t.Fatalf("consumption records wrong: %+v", records.List)
	}
}

func TestDisposalAPI_RejectAndEligibility(t *testing.T) {
	a := setupDisposalAPI(t)
	expiring := a.seedFood("临期面包", constants.FreshnessFresh, 1, time.Now().AddDate(0, 0, 1))
	fresh := a.seedFood("充裕苹果", constants.FreshnessFresh, 5, time.Now().AddDate(0, 0, 20))

	// 非临期/过期食品不能申请 → 409 / 1113
	if status, resp := a.do(http.MethodPost, "/api/v1/disposals", a.memberTok, map[string]any{
		"food_item_id": fresh.ID, "quantity": 1, "method": "consume",
	}); status != http.StatusConflict || resp.Code != constants.CodeDisposalNotEligible {
		t.Fatalf("fresh food: status=%d code=%d", status, resp.Code)
	}

	// 非法处置方式 → 400（validator oneof）
	if status, _ := a.do(http.MethodPost, "/api/v1/disposals", a.memberTok, map[string]any{
		"food_item_id": expiring.ID, "quantity": 1, "method": "burn",
	}); status != http.StatusBadRequest {
		t.Fatalf("invalid method status=%d, want 400", status)
	}

	// 超量申请 → 409 / 1102
	if status, resp := a.do(http.MethodPost, "/api/v1/disposals", a.memberTok, map[string]any{
		"food_item_id": expiring.ID, "quantity": 3, "method": "discard",
	}); status != http.StatusConflict || resp.Code != constants.CodeQuantityExceeds {
		t.Fatalf("exceeds: status=%d code=%d", status, resp.Code)
	}

	// 正常申请后驳回
	_, resp := a.do(http.MethodPost, "/api/v1/disposals", a.memberTok, map[string]any{
		"food_item_id": expiring.ID, "quantity": 1, "method": "donate",
	})
	var req model.DisposalRequest
	_ = json.Unmarshal(resp.Data, &req)

	status, resp := a.do(http.MethodPost, fmt.Sprintf("/api/v1/disposals/%d/reject", req.ID), a.adminTok, map[string]any{"remark": "自己吃"})
	if status != http.StatusOK || resp.Code != 0 {
		t.Fatalf("reject: status=%d code=%d msg=%s", status, resp.Code, resp.Msg)
	}
	var rejected model.DisposalRequest
	_ = json.Unmarshal(resp.Data, &rejected)
	if rejected.Status != constants.DisposalStatusRejected || rejected.ReviewRemark != "自己吃" {
		t.Fatalf("rejected request wrong: %+v", rejected)
	}

	// 驳回后库存不变，且再审批/驳回返回 409
	var food model.FoodItem
	if err := a.db.First(&food, expiring.ID).Error; err != nil {
		t.Fatal(err)
	}
	if food.Quantity != 1 || food.Status == constants.FreshnessConsumed {
		t.Fatalf("reject changed stock: %.2f %s", food.Quantity, food.Status)
	}
	if status, _ := a.do(http.MethodPost, fmt.Sprintf("/api/v1/disposals/%d/approve", req.ID), a.adminTok, map[string]any{}); status != http.StatusConflict {
		t.Fatalf("approve after reject status=%d, want 409", status)
	}

	// 未登录 → 401
	if status, _ := a.do(http.MethodGet, "/api/v1/disposals?family_id=1", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("no token status=%d, want 401", status)
	}
}
