package router

import (
	"github.com/gin-gonic/gin"
)

// registerConsumptionRoutes 消耗记录相关路由。
func registerConsumptionRoutes(g *gin.RouterGroup, h Handlers) {
	g.GET("/consumptions", h.Consumption.List)
	g.GET("/consumptions/analysis", h.Consumption.Analysis)
}
