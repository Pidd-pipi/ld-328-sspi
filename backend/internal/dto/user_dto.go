package dto

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,max=50"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest 更新资料请求。
type UpdateProfileRequest struct {
	Name   string `json:"name" binding:"max=50"`
	Avatar string `json:"avatar" binding:"max=255"`
}

// TokenResponse 登录/注册响应。
type TokenResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}
