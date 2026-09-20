package router

import "github.com/gin-gonic/gin"

// registerDisposalRoutes 临期处置申请相关路由（提交/查询需家庭成员；批准/驳回在 service 内强制家庭 Admin）。
func registerDisposalRoutes(g *gin.RouterGroup, h Handlers) {
	g.POST("/disposal-applications", h.Disposal.Create)
	g.GET("/disposal-applications", h.Disposal.List)
	g.GET("/disposal-applications/:id", h.Disposal.Detail)
	g.POST("/disposal-applications/:id/approve", h.Disposal.Approve)
	g.POST("/disposal-applications/:id/reject", h.Disposal.Reject)
}
