package router

import (
	"github.com/gin-gonic/gin"
)

// registerFamilyRoutes 家庭组相关路由。
func registerFamilyRoutes(g *gin.RouterGroup, h Handlers) {
	g.POST("/family-groups", h.FamilyGroup.Create)
	g.GET("/family-groups", h.FamilyGroup.ListMine)
	g.GET("/family-groups/:id", h.FamilyGroup.Detail)
	g.POST("/family-groups/join", h.FamilyGroup.Join)
	g.POST("/family-groups/:id/invite", h.FamilyGroup.Invite)
	g.PUT("/family-groups/:id/members/:memberId/role", h.FamilyGroup.SetMemberRole)
	g.DELETE("/family-groups/:id/members/:memberId", h.FamilyGroup.RemoveMember)
}
