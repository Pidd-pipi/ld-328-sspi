package router

import (
	"github.com/blueship581/cyfreshfood/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// corsMiddleware 跨域配置：来源白名单从 APP_CORS_ORIGINS 环境变量读取，生产不允许通配符。
func corsMiddleware(cfg config.Config) gin.HandlerFunc {
	allowOrigins := cfg.CORSOriginsList()
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"http://localhost:18628"}
	}
	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
	})
}
