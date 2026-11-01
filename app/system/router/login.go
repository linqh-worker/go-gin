package router

import (
	controller "go-gin/app/system/controller/auth"
	user "go-gin/app/system/controller/user"
	"go-gin/app/system/middleware"
	"go-gin/internal/httpx"
)

func RegisterLoginRoutes(r *httpx.RouterGroup) {
	captcha := r.Group("")
	captcha.Before(middleware.AuthRateLimit("captcha", 30))
	captcha.GET("/captcha", controller.LoginController.Captcha)
	login := r.Group("")
	login.Before(middleware.AuthRateLimit("login", 10))
	login.POST("/login", controller.LoginController.Login)
	r.GET("/me", user.UserController.Me)
	r.POST("/logout", user.UserController.Logout)
}
