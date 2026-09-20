package constants

// 统一错误码：0 表示成功；业务错误从 1000 起分段。
const (
	CodeOK                 = 0
	CodeBadRequest         = 1000
	CodeUnauthorized       = 1001
	CodeForbidden          = 1002
	CodeNotFound           = 1003
	CodeConflict           = 1004
	CodeValidationError    = 1005
	CodeRateLimited        = 1006
	CodeInternalError      = 1007
	CodeFoodNotAvailable   = 1101
	CodeQuantityExceeds    = 1102
	CodeDisposalNotAllowed = 1103 // 食品非临期/过期，不可发起处置申请
	CodeDisposalDuplicate  = 1104 // 同一食品已存在待处理申请
	CodeDisposalNotPending = 1105 // 申请已结案，不能重复审批
	CodeDisposalNotFound   = 1106 // 处置申请不存在
	CodeFamilyNotMember    = 1201
	CodeFamilyFull         = 1202
	CodeDuplicatePhone     = 1301
	CodeLoginFailed        = 1302
	CodeUserNotFound       = 1303
	CodeRecipeNotFound     = 1401
)
