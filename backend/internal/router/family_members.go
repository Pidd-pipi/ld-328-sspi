package router

import (
	"github.com/blueship581/cyfreshfood/internal/handler"
	"github.com/gin-gonic/gin"
)

// registerFamilyMemberRoutes 家庭成员相关路由（与家庭组共用处理器）。
func registerFamilyMemberRoutes(g *gin.RouterGroup, h handler.FamilyGroupHandler) {
	g.PUT("/family-groups/:id/members/:memberId/role", h.SetMemberRole)
	g.DELETE("/family-groups/:id/members/:memberId", h.RemoveMember)
}
