package middleware

import (
	"crypto/sha256"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-gin/internal/component/redisx"
	"go-gin/internal/errorx"
	"go-gin/internal/httpx"
	"net"
	"net/http"
)

var authLimitScript = redis.NewScript(`local n=redis.call('INCR',KEYS[1]); if n==1 then redis.call('EXPIRE',KEYS[1],60) end; return n`)

// AuthRateLimit 按直连IP限制每分钟请求数，不信任可伪造的转发头。
func AuthRateLimit(action string, limit int64) httpx.HandlerFunc {
	return func(ctx *httpx.Context) (any, error) {
		client := redisx.Client()
		if client == nil {
			return nil, errorx.NewServerError(http.StatusServiceUnavailable)
		}
		ip, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
		if err != nil {
			ip = ctx.Request.RemoteAddr
		}
		key := fmt.Sprintf("api:auth:limit:%s:%x", action, sha256.Sum256([]byte(ip)))
		n, err := authLimitScript.Run(ctx.Request.Context(), client, []string{key}).Int64()
		if err != nil {
			return nil, errorx.NewServerError(http.StatusServiceUnavailable)
		}
		if n > limit {
			ctx.Header("Retry-After", "60")
			return nil, errorx.NewServerError(http.StatusTooManyRequests)
		}
		return nil, nil
	}
}
