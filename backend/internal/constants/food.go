package constants

// FoodCategory 食品类别枚举：前端 constants/food.ts 与后端 constants/food.go 必须保持一致。
const (
	FoodCategoryFresh  = "fresh"  // 生鲜
	FoodCategoryDairy  = "dairy"  // 乳制品
	FoodCategoryCooked = "cooked" // 熟食
	FoodCategoryBakery = "bakery" // 烘焙
	FoodCategoryFrozen = "frozen" // 冷冻
	FoodCategoryOther  = "other"  // 其他
)

// FoodCategories 全部食品类别（看板筛选器/统计分组/表单下拉共用）。
var FoodCategories = []string{
	FoodCategoryFresh, FoodCategoryDairy, FoodCategoryCooked,
	FoodCategoryBakery, FoodCategoryFrozen, FoodCategoryOther,
}

// FreshnessStatus 新鲜度状态枚举：前端 constants/food.ts 与后端 constants/food.go 必须保持一致。
const (
	FreshnessFresh    = "fresh"    // 充裕
	FreshnessExpiring = "expiring" // 临期（3 天内）
	FreshnessExpired  = "expired"  // 已过期
	FreshnessConsumed = "consumed" // 已消耗
)

// FreshnessStatuses 全部新鲜度状态。
var FreshnessStatuses = []string{
	FreshnessFresh, FreshnessExpiring, FreshnessExpired, FreshnessConsumed,
}

// ExpiringThresholdDays 临期阈值（天），与前端 utils/calculateRemainingDays.ts 保持一致。
const ExpiringThresholdDays = 3

// StorageLocations 存放位置枚举。
const (
	StorageFridge   = "fridge"   // 冰箱
	StoragePantry   = "pantry"   // 储藏室
	StorageFreezer  = "freezer"  // 冷冻室
	StorageCounter  = "counter"  // 台面
	StorageOtherLoc = "other"    // 其他
)

// StorageLocationList 全部存放位置。
var StorageLocationList = []string{StorageFridge, StoragePantry, StorageFreezer, StorageCounter, StorageOtherLoc}
