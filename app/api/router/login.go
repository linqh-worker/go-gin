package router

import (
	controller "go-gin/app/api/controller/auth"
	"go-gin/app/api/middleware"
	"go-gin/internal/httpx"
)

func RegisterLoginRoutes(r *httpx.RouterGroup) {
	key := r.Group("")
	key.Before(middleware.AuthRateLimit("public_key", 30))
	key.GET("/auth/public-key", controller.LoginController.PublicKey)
	captcha := r.Group("")
	captcha.Before(middleware.AuthRateLimit("captcha", 30))
	captcha.GET("/captcha", controller.LoginController.Captcha)
	login := r.Group("")
	login.Before(middleware.AuthRateLimit("login", 10))
	login.POST("/login", controller.LoginController.Login)
	register := r.Group("")
	register.Before(middleware.AuthRateLimit("register", 10))
	register.POST("/register", controller.LoginController.Register)
}
