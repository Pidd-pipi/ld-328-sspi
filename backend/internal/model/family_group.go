package model

import "time"

// FamilyGroup 家庭组：拥有成员与食品。
type FamilyGroup struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:50;not null" json:"name"`
	OwnerID    uint      `gorm:"index;not null" json:"owner_id"`
	InviteCode string    `gorm:"size:16;uniqueIndex;not null" json:"invite_code"`
	CreatedAt  time.Time `json:"created_at"`
	Owner      User      `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
}
