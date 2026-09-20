package dto

// CreateFamilyRequest 创建家庭组请求。
type CreateFamilyRequest struct {
	Name string `json:"name" binding:"required,max=50"`
}

// InviteMemberRequest 邀请成员请求。
type InviteMemberRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

// JoinFamilyRequest 加入家庭组请求。
type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code" binding:"required,max=16"`
}

// UpdateMemberRoleRequest 设置成员角色请求。
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}
