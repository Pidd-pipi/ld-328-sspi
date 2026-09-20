package middleware

import (
	"net/http"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/gin-gonic/gin"
)

// RequireRole 校验用户角色（UserRole），无权限返回 403。
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get(RoleKey)
		roleStr, _ := role.(string)
		if !allowed[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    constants.CodeForbidden,
				"message": constants.MsgRoleForbidden,
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
