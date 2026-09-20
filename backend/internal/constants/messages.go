package constants

// messages.go 集中管理前端提示文案、后端返回文案与日志文案（屎山耦合点之一）。

const (
	MsgSuccess              = "ok"
	MsgLoginSuccess         = "登录成功"
	MsgRegisterSuccess      = "注册成功"
	MsgCreated              = "创建成功"
	MsgUpdated              = "更新成功"
	MsgDeleted              = "删除成功"
	MsgConsumed             = "消耗成功"
	MsgImported             = "CSV 导入成功"
	MsgFamilyInvited        = "成员已邀请加入家庭组"
	MsgFamilyJoined         = "已加入家庭组"
	MsgNotificationRead     = "通知已标记为已读"
	MsgScanTriggered        = "临期扫描已触发"
	MsgRecipeRecommended    = "食谱推荐已生成"
	MsgFoodExpired          = "食品已过期"
	MsgFoodExpiring         = "食品临近过期"
	MsgReportGenerated      = "统计报表已生成"
	MsgNotFamilyMember      = "您不是该家庭组（FamilyGroup）成员，无法执行此操作"
	MsgFoodConsumeFailed    = "食品（FoodItem）消耗失败：数量超出库存"
	MsgFoodNotAvailable     = "食品（FoodItem）已消耗或不存在，无法操作"
	MsgCategoryInvalid      = "食品类别（FoodCategory）不合法"
	MsgStatusInvalid        = "新鲜度状态（FreshnessStatus）不合法"
	MsgDuplicatePhone       = "手机号（User.phone）已注册"
	MsgLoginFailed          = "手机号或密码（User）错误"
	MsgRoleForbidden        = "当前角色（UserRole）无权执行该操作"
	MsgRateLimited          = "请求过于频繁，请稍后重试"
	MsgInternalError        = "服务内部错误"
	MsgParamInvalid         = "请求参数校验失败"
	MsgUploadInvalid        = "上传文件不合法"
)


// MsgUserNotFound 用户不存在文案。
const MsgUserNotFound = "用户（User）不存在"
