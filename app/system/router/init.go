package router

import (
	"go-gin/app/system/middleware"
	"go-gin/internal/httpx"
)

// Register 后台账号认证独立于用户端；不开放管理员注册。
func Register(r *httpx.RouterGroup) {
	r.Use(middleware.Module())
	r.Before(middleware.Authorize())
	RegisterLoginRoutes(r)
	RegisterRBACRoutes(r)
	RegisterMenuRoutes(r)
	RegisterAccountRoutes(r)
	RegisterConfigRoutes(r)

}
