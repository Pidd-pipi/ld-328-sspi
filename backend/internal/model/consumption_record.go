package model

import "time"

// ConsumptionRecord 消耗记录：记录食品消耗与操作人。
type ConsumptionRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FoodItemID uint      `gorm:"index;not null" json:"food_item_id"`
	Quantity   float64   `json:"quantity"`
	ConsumedAt time.Time `json:"consumed_at"`
	UserID     uint      `gorm:"index" json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	User       User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FoodItem   FoodItem  `gorm:"foreignKey:FoodItemID" json:"food_item,omitempty"`
}
