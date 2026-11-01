package httpx

import (
	"github.com/gin-gonic/gin"
	"go-gin/internal/component/logx"
	"strings"
)

// ModuleLog 在鉴权、数据库检查及404前标记请求，确保拒绝和panic也正确分流。
func ModuleLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		module := logx.DefaultModule()
		switch {
		case path == "/api" || strings.HasPrefix(path, "/api/"):
			module = "api"
		case path == "/system" || strings.HasPrefix(path, "/system/") || path == "/monitor" || strings.HasPrefix(path, "/monitor/"):
			module = "system"
		}
		c.Set("module", module)
		c.Request = c.Request.WithContext(logx.WithModule(c.Request.Context(), module))
		c.Next()
	}
}
