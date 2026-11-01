package middleware

import (
	"github.com/gin-gonic/gin"
	"go-gin/internal/component/logx"
)

// Module 只标记模块，不承担管理员认证或权限校验。
func Module() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("module", "system")
		c.Request = c.Request.WithContext(logx.WithModule(c.Request.Context(), "system"))
		c.Next()
	}
}
