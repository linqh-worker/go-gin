package router

import (
	"go-gin/app/api/middleware"
	"go-gin/internal/httpx"
)

// Register 集中注册用户端路由，所有路径都位于 /api 下。
func Register(r *httpx.RouterGroup) {
	r.Use(middleware.Module())
	RegisterApiRoutes(r)
	RegisterLoginRoutes(r)
	RegisterUserRoutes(r)
}
