package middleware

import "github.com/gin-gonic/gin"

// Module 标记请求所属模块。用户端认证中间件也放在这个包中。
func Module() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("module", "api")
		c.Next()
	}
}
