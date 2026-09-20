package model

import "time"

// Recipe 简易食谱：按食品类别推荐搭配。
type Recipe struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"size:100;not null" json:"name"`
	Ingredients      string    `gorm:"size:1000" json:"ingredients"`
	Description      string    `gorm:"size:2000" json:"description"`
	SuitableCategory string    `gorm:"size:20;index" json:"suitable_category"`
	CreatedAt        time.Time `json:"created_at"`
}
