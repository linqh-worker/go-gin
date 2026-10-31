package router

import (
	controller "go-gin/app/api/controller/user"
	"go-gin/app/api/middleware"
	"go-gin/internal/httpx"
)

func RegisterUserRoutes(r *httpx.RouterGroup) {
	// r 是 /api；认证只挂在该子组，不影响公开的登录、注册和验证码。
	private := r.Group("")
	private.Before(middleware.TokenCheck())
	//用户接口
	userC := controller.UserController
	private.GET("/me", userC.Me)
	private.POST("/logout", userC.LoginOut)
	// 原有列表、添加、批量添加只是演示代码，不向普通用户开放全表管理能力。
}
