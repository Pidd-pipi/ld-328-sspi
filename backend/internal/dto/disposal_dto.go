package dto

// CreateDisposalRequest 提交临期处置申请请求。
type CreateDisposalRequest struct {
	FoodItemID uint    `json:"food_item_id" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	Method     string  `json:"method" binding:"required,oneof=discard eat donate"`
}

// ReviewDisposalRequest 批准/驳回处置申请请求（仅家庭管理员）。
type ReviewDisposalRequest struct {
	Note string `json:"note" binding:"max=255"`
}
