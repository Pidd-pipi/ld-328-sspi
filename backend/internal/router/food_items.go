package router

import (
	"github.com/gin-gonic/gin"
)

// registerFoodRoutes 食品相关路由。
func registerFoodRoutes(g *gin.RouterGroup, h Handlers) {
	g.POST("/foods", h.FoodItem.Create)
	g.GET("/foods", h.FoodItem.List)
	g.POST("/foods/csv-import", h.FoodItem.CSVImport)
	g.GET("/foods/:id", h.FoodItem.Detail)
	g.PUT("/foods/:id", h.FoodItem.Update)
	g.DELETE("/foods/:id", h.FoodItem.Delete)
	g.POST("/foods/:id/consume", h.FoodItem.Consume)
}
