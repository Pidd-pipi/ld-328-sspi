package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/util"
	"github.com/gin-gonic/gin"
)

// ErrorHandler 全局错误处理：恢复 panic 并统一响应格式。
func ErrorHandler(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error(constants.LOG_PANIC_RECOVERED, "panic", rec, "path", c.Request.URL.Path)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code": constants.CodeInternalError, "message": constants.MsgInternalError, "data": nil,
				})
			}
		}()
		c.Next()
		// 已写响应则跳过
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		err := c.Errors.Last().Err
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{"code": appErr.Code, "message": appErr.Message, "data": nil})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": constants.CodeInternalError, "message": constants.MsgInternalError, "data": nil,
		})
	}
}
