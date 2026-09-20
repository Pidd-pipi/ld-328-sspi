package router

import (
	"github.com/blueship581/cyfreshfood/internal/handler"
	"github.com/gin-gonic/gin"
)

// registerUserRoutes 用户相关路由（在 router.go 中已注册，此文件保留实体路由入口）。
func registerUserRoutes(g *gin.RouterGroup, h handler.UserHandler) {
	g.GET("/users/me", h.Me)
	g.PUT("/users/me", h.UpdateProfile)
}
