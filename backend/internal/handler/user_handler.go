package handler

import (
	"errors"
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/dto"
	"github.com/blueship581/cyfreshfood/internal/middleware"
	"github.com/blueship581/cyfreshfood/internal/service"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户接口。
type UserHandler struct {
	svc *service.UserService
	log *slog.Logger
}

// NewUserHandler 构造用户接口。
func NewUserHandler(svc *service.UserService, log *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

// Register 注册。
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("注册参数（User）不合法", err))
		return
	}
	user, token, err := h.svc.Register(c.Request.Context(), req.Phone, req.Password, req.Name)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, dto.TokenResponse{Token: token, User: user})
}

// Login 登录。
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("登录参数（User）不合法", err))
		return
	}
	user, token, err := h.svc.Login(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.TokenResponse{Token: token, User: user})
}

// Me 当前用户信息。
func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.Error(util.UnauthorizedError("未登录（User token）", errors.New("no user id")))
		return
	}
	user, err := h.svc.GetByID(c.Request.Context(), userID.(uint))
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, user)
}

// UpdateProfile 修改资料。
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get(middleware.UserIDKey)
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("资料参数（User）不合法", err))
		return
	}
	user, err := h.svc.UpdateProfile(c.Request.Context(), userID.(uint), req.Name, req.Avatar)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, user)
}
