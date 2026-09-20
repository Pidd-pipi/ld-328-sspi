package model

// CategoryCount 按类别统计结果（看板/统计报表共用）。
type CategoryCount struct {
	Category      string  `json:"category"`
	Count         int64   `json:"count"`
	TotalQuantity float64 `json:"total_quantity"`
}

// MonthlyConsumption 按月消耗统计。
type MonthlyConsumption struct {
	Category      string  `json:"category"`
	Count         int64   `json:"count"`
	TotalQuantity float64 `json:"total_quantity"`
}

// TopFood 排行（最常购买/最常浪费）。
type TopFood struct {
	FoodItemID uint    `json:"food_item_id"`
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Quantity   float64 `json:"quantity"`
}
