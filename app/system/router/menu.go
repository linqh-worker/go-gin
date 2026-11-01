package router

import (
	controller "go-gin/app/system/controller/menu"
	"go-gin/internal/httpx"
)

func RegisterMenuRoutes(r *httpx.RouterGroup) {
	r.GET("/menus/mine", controller.Mine)
	r.GET("/menus", controller.Tree)
	r.POST("/menus", controller.Save)
	r.PUT("/menus/:id", controller.Save)
	r.DELETE("/menus/:id", controller.Delete)
	r.GET("/roles/:id/menus", controller.RoleMenus)
	r.PUT("/roles/:id/menus", controller.SetRoleMenus)
}
