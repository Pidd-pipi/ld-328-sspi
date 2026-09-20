package model

import "time"

// FoodItem 食品：隶属家庭组，记录保质期与新鲜度状态。
type FoodItem struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	FamilyID        uint       `gorm:"index;not null" json:"family_id"`
	Name            string     `gorm:"size:100;not null" json:"name"`
	Category        string     `gorm:"size:20;not null" json:"category"`
	ProductionDate  *time.Time `json:"production_date"`
	ShelfLifeDays   int        `json:"shelf_life_days"`
	Quantity        float64    `json:"quantity"`
	Unit            string     `gorm:"size:20" json:"unit"`
	StorageLocation string     `gorm:"size:20" json:"storage_location"`
	OpenedAt        *time.Time `json:"opened_at"`
	ExpiryDate      *time.Time `json:"expiry_date"`
	Status          string     `gorm:"size:20;default:fresh;index" json:"status"`
	ImageURL        string     `gorm:"size:255" json:"image_url"`
	CreatorID       uint       `json:"creator_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
