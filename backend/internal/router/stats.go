package router

import (
	"github.com/gin-gonic/gin"
)

// registerStatsRoutes 统计相关路由。
func registerStatsRoutes(g *gin.RouterGroup, h Handlers) {
	g.GET("/stats/dashboard", h.Stats.Dashboard)
	g.GET("/stats/statistics", h.Stats.Statistics)
	g.GET("/stats/statistics/export", h.Stats.ExportPDF)
}
