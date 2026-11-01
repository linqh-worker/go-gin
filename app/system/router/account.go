package router

import (
	controller "go-gin/app/system/controller/account"
	"go-gin/internal/httpx"
)

func RegisterAccountRoutes(r *httpx.RouterGroup) {
	r.GET("/users", controller.List(false))
	r.GET("/users/:id", controller.Detail(false))
	r.POST("/users", controller.Create(false))
	r.PATCH("/users/:id", controller.Update(false))
	r.PUT("/users/:id/status", controller.Status(false))
	r.GET("/admins", controller.List(true))
	r.GET("/admins/:id", controller.Detail(true))
	r.POST("/admins", controller.Create(true))
	r.PATCH("/admins/:id", controller.Update(true))
	r.PUT("/admins/:id/status", controller.Status(true))
}
