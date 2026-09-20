package router

import (
	"github.com/gin-gonic/gin"
)

// registerNotificationRoutes 通知相关路由。
func registerNotificationRoutes(g *gin.RouterGroup, h Handlers) {
	g.GET("/notifications", h.Notification.List)
	g.GET("/notifications/unread-count", h.Notification.UnreadCount)
	g.PUT("/notifications/read-all", h.Notification.MarkAllRead)
	g.PUT("/notifications/:id/read", h.Notification.MarkRead)
}
