package router

import "github.com/gin-gonic/gin"

// registerDisposalRoutes 临期处置申请相关路由。
// 提交/查看：家庭成员；批准/驳回：仅家庭管理员（service 层按 FamilyMember.role 校验）。
func registerDisposalRoutes(g *gin.RouterGroup, h Handlers) {
	g.POST("/disposals", h.Disposal.Create)
	g.GET("/disposals", h.Disposal.List)
	g.GET("/disposals/:id", h.Disposal.Detail)
	g.POST("/disposals/:id/approve", h.Disposal.Approve)
	g.POST("/disposals/:id/reject", h.Disposal.Reject)
}
