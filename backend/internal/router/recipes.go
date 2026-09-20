package router

import (
	"github.com/gin-gonic/gin"
)

// registerRecipeRoutes 食谱相关路由。
func registerRecipeRoutes(g *gin.RouterGroup, h Handlers) {
	g.GET("/recipes/recommendations", h.Recipe.Recommendations)
	g.GET("/recipes", h.Recipe.List)
	g.GET("/recipes/:id", h.Recipe.Detail)
}
