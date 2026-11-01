package httpx

import (
	"github.com/gin-gonic/gin"
	"go-gin/internal/component/logx"
	"time"
)

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		cost := time.Since(start)
		if cost > time.Minute {
			cost = cost.Truncate(time.Second)
		}
		// 只记录路径，不记录可能包含密码、验证码或Token的查询参数。
		logx.AccessLogger.Info().Ctx(c.Request.Context()).Str("path", path).Str("method", c.Request.Method).Str("ip", c.ClientIP()).Str("cost", cost.String()).Int("status", c.Writer.Status()).Str("proto", c.Request.Proto).Str("user_agent", c.Request.UserAgent()).Send()
	}
}
