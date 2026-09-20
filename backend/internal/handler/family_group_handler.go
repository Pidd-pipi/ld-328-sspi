package handler

import (
	"log/slog"
	"strconv"

	"github.com/blueship581/cyfreshfood/internal/dto"
	"github.com/blueship581/cyfreshfood/internal/middleware"
	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// FamilyGroupHandler 家庭组接口。
type FamilyGroupHandler struct {
	svc      *service.FamilyGroupService
	memberSvc *service.FamilyMemberService
	log      *slog.Logger
}

// NewFamilyGroupHandler 构造家庭组接口。
func NewFamilyGroupHandler(svc *service.FamilyGroupService, memberSvc *service.FamilyMemberService, log *slog.Logger) *FamilyGroupHandler {
	return &FamilyGroupHandler{svc: svc, memberSvc: memberSvc, log: log}
}

func userID(c *gin.Context) uint {
	v, _ := c.Get(middleware.UserIDKey)
	return v.(uint)
}

// Create 创建家庭组。
func (h *FamilyGroupHandler) Create(c *gin.Context) {
	var req dto.CreateFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("家庭组（FamilyGroup）参数不合法", err))
		return
	}
	group, err := h.svc.Create(c.Request.Context(), userID(c), req.Name)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, group)
}

// ListMine 我的家庭组。
func (h *FamilyGroupHandler) ListMine(c *gin.Context) {
	groups, err := h.svc.ListMine(c.Request.Context(), userID(c))
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, groups)
}

// Detail 家庭组详情（含成员）。
func (h *FamilyGroupHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	group, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	if err := h.svc.IsMember(c.Request.Context(), id, userID(c)); err != nil {
		c.Error(err)
		return
	}
	members, err := h.memberSvc.ListByFamily(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"group": group, "members": members})
}

// Invite 邀请成员。
func (h *FamilyGroupHandler) Invite(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("邀请成员（FamilyMember）参数不合法", err))
		return
	}
	member, err := h.svc.InviteMember(c.Request.Context(), id, userID(c), req.UserID)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, member)
}

// Join 通过邀请码加入。
func (h *FamilyGroupHandler) Join(c *gin.Context) {
	var req dto.JoinFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("加入家庭组（FamilyGroup）参数不合法", err))
		return
	}
	group, err := h.svc.JoinByCode(c.Request.Context(), userID(c), req.InviteCode)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, group)
}

// SetMemberRole 设置成员权限。
func (h *FamilyGroupHandler) SetMemberRole(c *gin.Context) {
	familyID := parseUint(c.Param("id"))
	memberID := parseUint(c.Param("memberId"))
	var req dto.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("成员角色（FamilyMember.role）不合法", err))
		return
	}
	if err := h.svc.SetMemberRole(c.Request.Context(), familyID, userID(c), memberID, req.Role); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"member_id": memberID, "role": req.Role})
}

// RemoveMember 移除成员。
func (h *FamilyGroupHandler) RemoveMember(c *gin.Context) {
	familyID := parseUint(c.Param("id"))
	memberID := parseUint(c.Param("memberId"))
	if err := h.svc.RemoveMember(c.Request.Context(), familyID, userID(c), memberID); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"removed": memberID})
}

func parseUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}
