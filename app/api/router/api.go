package router

import (
	"go-gin/internal/httpx"
)

func RegisterApiRoutes(r *httpx.RouterGroup) {
	r.GET("/", func(ctx *httpx.Context) (any, error) {
		return map[string]string{"module": "api"}, nil
	})
}
