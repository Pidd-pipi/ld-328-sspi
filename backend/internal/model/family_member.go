package model

import "time"

// FamilyMember 家庭成员：关联家庭组与用户，含家庭内角色。
type FamilyMember struct {
	ID       uint        `gorm:"primaryKey" json:"id"`
	FamilyID uint        `gorm:"index;not null" json:"family_id"`
	UserID   uint        `gorm:"index;not null" json:"user_id"`
	Role     string      `gorm:"size:20;default:member" json:"role"`
	JoinedAt time.Time   `json:"joined_at"`
	User     User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Family   FamilyGroup `gorm:"foreignKey:FamilyID" json:"family,omitempty"`
}
