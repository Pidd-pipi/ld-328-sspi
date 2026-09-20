package constants

// DisposalMethod 临期处置方式枚举：前端 constants/food.ts 与后端 constants/disposal.go 必须保持一致。
const (
	DisposalMethodDiscard = "discard" // 丢弃
	DisposalMethodEat     = "eat"     // 食用
	DisposalMethodDonate  = "donate"  // 捐赠
)

// DisposalMethods 全部处置方式（申请表单/日志模板/formatters 共用）。
var DisposalMethods = []string{DisposalMethodDiscard, DisposalMethodEat, DisposalMethodDonate}

// DisposalStatus 处置申请状态枚举。
const (
	DisposalStatusPending  = "pending"  // 待处理
	DisposalStatusApproved = "approved" // 已批准（结案）
	DisposalStatusRejected = "rejected" // 已驳回（关闭）
)

// DisposalStatuses 全部处置申请状态。
var DisposalStatuses = []string{DisposalStatusPending, DisposalStatusApproved, DisposalStatusRejected}
