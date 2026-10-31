package router

import (
	controller "go-gin/app/system/controller/menu"
	"go-gin/app/system/middleware"
	"go-gin/internal/httpx"
)

func RegisterMenuRoutes(r *httpx.RouterGroup) {
	private := r.Group("")
	private.Before(middleware.TokenCheck())
	private.GET("/menus/mine", controller.Mine)
	manage := private.Group("")
	manage.Before(middleware.RequireSuper())
	manage.GET("/menus", controller.Tree)
	manage.POST("/menus", controller.Save)
	manage.PUT("/menus/:id", controller.Save)
	manage.DELETE("/menus/:id", controller.Delete)
	manage.GET("/roles/:id/menus", controller.RoleMenus)
	manage.PUT("/roles/:id/menus", controller.SetRoleMenus)
}
