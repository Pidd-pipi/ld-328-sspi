package model

import "time"

// DisposalApplication 临期处置申请：家庭成员对临期/过期食品发起处置闭环。
// 同一 FoodItem 只允许存在一张 pending 待处理申请（数据库部分唯一索引兜底，见 repository）。
type DisposalApplication struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	FamilyID    uint    `gorm:"index;not null" json:"family_id"`
	FoodItemID  uint    `gorm:"index;not null" json:"food_item_id"`
	ApplicantID uint    `gorm:"index;not null" json:"applicant_id"`
	Quantity    float64 `gorm:"not null" json:"quantity"`
	Method      string  `gorm:"size:20;not null" json:"method"`
	// Status: pending 待处理 / approved 已批准 / rejected 已驳回（终态）。
	Status     string     `gorm:"size:20;default:pending;index" json:"status"`
	ReviewerID *uint      `gorm:"index" json:"reviewer_id"`
	ReviewNote string     `gorm:"size:255" json:"review_note"`
	ReviewedAt *time.Time `json:"reviewed_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Applicant  User       `gorm:"foreignKey:ApplicantID" json:"applicant,omitempty"`
	Reviewer   *User      `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	FoodItem   FoodItem   `gorm:"foreignKey:FoodItemID" json:"food_item,omitempty"`
}
