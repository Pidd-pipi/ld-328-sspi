package router

import (
	"log/slog"
	"net/http"

	"github.com/blueship581/cyfreshfood/internal/config"
	"github.com/blueship581/cyfreshfood/internal/handler"
	"github.com/blueship581/cyfreshfood/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Handlers 全部接口处理器集合。
type Handlers struct {
	User          *handler.UserHandler
	FamilyGroup   *handler.FamilyGroupHandler
	FoodItem      *handler.FoodItemHandler
	Consumption   *handler.ConsumptionRecordHandler
	Notification  *handler.NotificationHandler
	Recipe        *handler.RecipeHandler
	Stats         *handler.StatsHandler
}

// New 装配 Gin 路由。
func New(cfg config.Config, log *slog.Logger, h Handlers, limiter *middleware.RateLimiter) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.ErrorHandler(log))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(log))
	r.Use(corsMiddleware(cfg))

	auth := middleware.AuthRequired(cfg.JWTSecret)
	rate := limiter.Limit()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "up"}})
	})

	api := r.Group("/api/v1")
	api.POST("/auth/register", rate, h.User.Register)
	api.POST("/auth/login", rate, h.User.Login)

	authed := api.Group("")
	authed.Use(auth)
	authed.Use(rate)
	{
		authed.GET("/users/me", h.User.Me)
		authed.PUT("/users/me", h.User.UpdateProfile)
	}

	registerFamilyRoutes(authed, h)
	registerFoodRoutes(authed, h)
	registerConsumptionRoutes(authed, h)
	registerNotificationRoutes(authed, h)
	registerRecipeRoutes(authed, h)
	registerStatsRoutes(authed, h)
	return r
}
