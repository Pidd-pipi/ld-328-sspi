package dto

import "time"

// CreateFoodRequest 创建/更新食品请求。
type CreateFoodRequest struct {
	FamilyID        uint       `json:"family_id" binding:"required"`
	Name            string     `json:"name" binding:"required,max=100"`
	Category        string     `json:"category" binding:"required"`
	ProductionDate  *time.Time `json:"production_date"`
	ShelfLifeDays   int        `json:"shelf_life_days"`
	Quantity        float64    `json:"quantity" binding:"gte=0"`
	Unit            string     `json:"unit"`
	StorageLocation string     `json:"storage_location"`
	OpenedAt        *time.Time `json:"opened_at"`
	ImageURL        string     `json:"image_url"`
}

// CSVImportRequest CSV 批量导入请求。
type CSVImportRequest struct {
	FamilyID uint   `json:"family_id" binding:"required"`
	CSVText  string `json:"csv_text" binding:"required"`
}

// ConsumeRequest 消耗请求。
type ConsumeRequest struct {
	Quantity   float64   `json:"quantity" binding:"required,gt=0"`
	ConsumedAt time.Time `json:"consumed_at"`
}
