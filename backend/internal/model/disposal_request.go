package model

import "time"

// DisposalRequest 临期处置申请：家庭成员对临期/过期食品发起，家庭管理员审批闭环。
// 处置方式（discard/consume/donate）与状态（pending/approved/rejected）枚举统一定义于
// constants/food.go；同一 FoodItem 仅允许存在一条 pending 申请（数据库部分唯一索引兜底）。
type DisposalRequest struct {
	ID                  uint               `gorm:"primaryKey" json:"id"`
	FamilyID            uint               `gorm:"index;not null" json:"family_id"`
	FoodItemID          uint               `gorm:"index;not null" json:"food_item_id"`
	ApplicantID         uint               `gorm:"index;not null" json:"applicant_id"`
	Quantity            float64            `gorm:"not null" json:"quantity"`
	Method              string             `gorm:"size:20;not null" json:"method"`
	Reason              string             `gorm:"size:255" json:"reason"`
	Status              string             `gorm:"size:20;default:pending;index" json:"status"`
	ReviewerID          *uint              `gorm:"index" json:"reviewer_id"`
	ReviewRemark        string             `gorm:"size:255" json:"review_remark"`
	ConsumptionRecordID *uint              `gorm:"index" json:"consumption_record_id"`
	ReviewedAt          *time.Time         `json:"reviewed_at"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	FoodItem            FoodItem           `gorm:"foreignKey:FoodItemID" json:"food_item,omitempty"`
	Applicant           User               `gorm:"foreignKey:ApplicantID" json:"applicant,omitempty"`
	Reviewer            *User              `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	ConsumptionRecord   *ConsumptionRecord `gorm:"foreignKey:ConsumptionRecordID" json:"consumption_record,omitempty"`
}
