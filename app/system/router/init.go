package router

import (
	controller "go-gin/app/system/controller/home"
	"go-gin/app/system/middleware"
	"go-gin/internal/httpx"
)

// Register 注册后台模块。当前仅提供公开的存活检查。
func Register(r *httpx.RouterGroup) {
	r.Use(middleware.Module())
	r.GET("/ping", controller.IndexController.Ping)
	// 后台登录、管理员认证和权限校验实现后，再注册受保护的业务接口。
}
