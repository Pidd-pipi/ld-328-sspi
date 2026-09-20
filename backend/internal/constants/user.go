package constants

// UserRole 用户角色枚举：前端 constants/user.ts 与后端 constants/user.go 必须保持一致。
const (
	RoleAdmin  = "admin"  // 管理员
	RoleMember = "member" // 普通成员
)

// Roles 全部角色。
var Roles = []string{RoleAdmin, RoleMember}

// FamilyMemberRole 家庭成员角色（家庭内权限）。
const (
	FamilyRoleAdmin  = "admin"
	FamilyRoleMember = "member"
)

// DefaultAvatar 默认头像地址。
const DefaultAvatar = "/default-avatar.png"
