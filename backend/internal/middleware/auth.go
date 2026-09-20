package middleware

import (
	"net/http"
	"strings"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// UserKey / RoleKey / UserIDKey 注入 gin.Context 的键。
const (
	UserKey   = "user"
	RoleKey   = "userRole"
	UserIDKey = "userID"
)

// AuthRequired 校验 Bearer JWT 并注入用户信息。
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			abortWithCode(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgLoginFailed)
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := util.ParseToken(jwtSecret, token)
		if err != nil {
			abortWithCode(c, http.StatusUnauthorized, constants.CodeUnauthorized, "登录状态已失效（User token）")
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set(RoleKey, claims.Role)
		c.Set(UserKey, claims)
		c.Next()
	}
}

func abortWithCode(c *gin.Context, status, code int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"code": code, "message": message, "data": nil})
}
