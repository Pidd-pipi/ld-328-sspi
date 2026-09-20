package model

import "time"

// Notification 提醒通知：临期/过期提醒。
type Notification struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	FamilyID   uint       `gorm:"index" json:"family_id"`
	FoodItemID uint       `gorm:"index" json:"food_item_id"`
	Type       string     `gorm:"size:20" json:"type"`
	Title      string     `gorm:"size:100" json:"title"`
	Content    string     `gorm:"size:500" json:"content"`
	IsRead     bool       `gorm:"default:false" json:"is_read"`
	SendAt     time.Time  `json:"send_at"`
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `json:"created_at"`
	FoodItem   FoodItem   `gorm:"foreignKey:FoodItemID" json:"food_item,omitempty"`
}
