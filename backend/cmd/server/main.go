package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
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
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}
	log := util.NewLogger()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		panic(fmt.Errorf("open database: %w", err))
	}
	if err := migrateAndSeed(db, log); err != nil {
		panic(fmt.Errorf("migrate database: %w", err))
	}
	log.Info(constants.LOG_DB_INITIALIZED)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr(), Password: cfg.RedisPassword})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Warn(constants.LOG_REDIS_CONNECTED, "error", err)
	} else {
		log.Info(constants.LOG_REDIS_CONNECTED, "addr", cfg.RedisAddr())
	}

	calculator := util.NewFoodCalculator()
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewFamilyGroupRepository(db)
	memberRepo := repository.NewFamilyMemberRepository(db)
	foodRepo := repository.NewFoodItemRepository(db)
	consumeRepo := repository.NewConsumptionRecordRepository(db)
	notifyRepo := repository.NewNotificationRepository(db)
	recipeRepo := repository.NewRecipeRepository(db)

	userSvc := service.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpireHours, log)
	familySvc := service.NewFamilyGroupService(groupRepo, memberRepo, log)
	memberSvc := service.NewFamilyMemberService(memberRepo, log)
	foodSvc := service.NewFoodItemService(foodRepo, consumeRepo, familySvc, calculator, log)
	consumeSvc := service.NewConsumptionRecordService(consumeRepo, foodRepo, familySvc, log)
	notifySvc := service.NewNotificationService(notifyRepo, familySvc, log)
	recipeSvc := service.NewRecipeService(recipeRepo, foodRepo, familySvc, calculator, log)
	statsSvc := service.NewStatsService(foodRepo, consumeRepo, notifyRepo, familySvc, memberSvc, calculator, log)
	reminderSvc := service.NewReminderService(foodRepo, notifyRepo, calculator, log)
	_ = service.NewNotificationSender(log)

	// 启动临期扫描定时任务（Redis 分布式锁）
	scheduler := service.NewReminderScheduler(reminderSvc, rdb, cfg.ScanInterval(), log)
	scheduler.Start(context.Background())
	defer scheduler.Stop()

	limiter := middleware.NewRateLimiter(cfg.RateLimitPerMin)
	h := router.Handlers{
		User:         handler.NewUserHandler(userSvc, log),
		FamilyGroup:  handler.NewFamilyGroupHandler(familySvc, memberSvc, log),
		FoodItem:     handler.NewFoodItemHandler(foodSvc, consumeSvc, log),
		Consumption:  handler.NewConsumptionRecordHandler(consumeSvc, log),
		Notification: handler.NewNotificationHandler(notifySvc, log),
		Recipe:       handler.NewRecipeHandler(recipeSvc, log),
		Stats:        handler.NewStatsHandler(statsSvc, log),
	}
	r := router.New(cfg, log, h, limiter)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Info(constants.LOG_SERVER_STARTED, "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(constants.LOG_SERVER_STARTED, "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error(constants.LOG_SERVER_SHUTDOWN, "error", err)
	}
}

// migrateAndSeed 自动迁移并注入种子数据。
// 若 database/init.sql 已建表（容器首次启动自动执行），则跳过 AutoMigrate，避免约束名冲突。
func migrateAndSeed(db *gorm.DB, log *slog.Logger) error {
	var tableCount int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='users'").Scan(&tableCount).Error; err != nil {
		return err
	}
	if tableCount > 0 {
		return nil // init.sql 已初始化
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.FamilyGroup{}, &model.FamilyMember{},
		&model.FoodItem{}, &model.ConsumptionRecord{}, &model.Notification{}, &model.Recipe{},
	); err != nil {
		return err
	}
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}
	seeds := []struct {
		phone    string
		password string
		name     string
		role     string
	}{
		{phone: "13800000001", password: "admin123", name: "管理员", role: constants.RoleAdmin},
		{phone: "13800000002", password: "member123", name: "家庭成员", role: constants.RoleMember},
	}
	users := make([]model.User, 0, len(seeds))
	for _, s := range seeds {
		hash, err := bcrypt.GenerateFromPassword([]byte(s.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash seed password: %w", err)
		}
		users = append(users, model.User{Phone: s.phone, PasswordHash: string(hash), Name: s.name, Role: s.role})
	}
	if err := db.Create(&users).Error; err != nil {
		return err
	}
	group := model.FamilyGroup{Name: "幸福之家", OwnerID: users[0].ID, InviteCode: "FAMILY01"}
	if err := db.Create(&group).Error; err != nil {
		return err
	}
	owner := model.FamilyMember{FamilyID: group.ID, UserID: users[0].ID, Role: constants.FamilyRoleAdmin}
	member := model.FamilyMember{FamilyID: group.ID, UserID: users[1].ID, Role: constants.FamilyRoleMember}
	if err := db.Create(&[]model.FamilyMember{owner, member}).Error; err != nil {
		return err
	}
	now := time.Now()
	foods := []model.FoodItem{
		{FamilyID: group.ID, Name: "鲜牛奶", Category: constants.FoodCategoryDairy, Quantity: 2, Unit: "盒", ShelfLifeDays: 5, StorageLocation: constants.StorageFridge, Status: constants.FreshnessFresh, CreatorID: users[0].ID, ExpiryDate: ptrTime(now.AddDate(0, 0, 2))},
		{FamilyID: group.ID, Name: "吐司面包", Category: constants.FoodCategoryBakery, Quantity: 1, Unit: "袋", ShelfLifeDays: 3, StorageLocation: constants.StoragePantry, Status: constants.FreshnessFresh, CreatorID: users[0].ID, ExpiryDate: ptrTime(now.AddDate(0, 0, 1))},
		{FamilyID: group.ID, Name: "鸡胸肉", Category: constants.FoodCategoryFresh, Quantity: 3, Unit: "块", ShelfLifeDays: 10, StorageLocation: constants.StorageFreezer, Status: constants.FreshnessFresh, CreatorID: users[1].ID, ExpiryDate: ptrTime(now.AddDate(0, 0, 7))},
		{FamilyID: group.ID, Name: "熟食卤味", Category: constants.FoodCategoryCooked, Quantity: 1, Unit: "份", ShelfLifeDays: 2, StorageLocation: constants.StorageFridge, Status: constants.FreshnessFresh, CreatorID: users[1].ID, ExpiryDate: ptrTime(now.AddDate(0, 0, -1))},
	}
	if err := db.Create(&foods).Error; err != nil {
		return err
	}
	if err := db.Create(&model.ConsumptionRecord{FoodItemID: foods[1].ID, Quantity: 1, UserID: users[0].ID, ConsumedAt: now.Add(-24 * time.Hour)}).Error; err != nil {
		return err
	}
	if err := db.Create(&[]model.Notification{
		{FamilyID: group.ID, FoodItemID: foods[0].ID, Type: constants.NotificationExpiring, Title: "食品临近过期", Content: "鲜牛奶 即将过期，请及时处理。", IsRead: false, SendAt: now},
		{FamilyID: group.ID, FoodItemID: foods[3].ID, Type: constants.NotificationExpired, Title: "食品已过期", Content: "熟食卤味 已过期，请及时处理。", IsRead: false, SendAt: now},
	}).Error; err != nil {
		return err
	}
	recipes := []model.Recipe{
		{Name: "牛奶燕麦粥", Ingredients: "牛奶、燕麦、蜂蜜", Description: "将燕麦煮开后加入牛奶与蜂蜜，营养早餐。", SuitableCategory: constants.FoodCategoryDairy},
		{Name: "蒜香吐司", Ingredients: "吐司、黄油、蒜末", Description: "吐司抹黄油蒜末烤至金黄。", SuitableCategory: constants.FoodCategoryBakery},
		{Name: "香煎鸡胸", Ingredients: "鸡胸肉、黑胡椒、橄榄油", Description: "鸡胸肉腌制后煎熟，低脂高蛋白。", SuitableCategory: constants.FoodCategoryFresh},
		{Name: "凉拌卤味", Ingredients: "卤味、黄瓜、香菜", Description: "卤味切片配黄瓜香菜凉拌。", SuitableCategory: constants.FoodCategoryCooked},
	}
	if err := db.Create(&recipes).Error; err != nil {
		return err
	}
	log.Info(constants.LOG_DB_INITIALIZED, "seed", "ok")
	return nil
}

func ptrTime(t time.Time) *time.Time { return &t }
