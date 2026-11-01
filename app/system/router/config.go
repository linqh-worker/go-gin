package router

import (
	controller "go-gin/app/system/controller/config"
	"go-gin/internal/httpx"
)

func RegisterConfigRoutes(r *httpx.RouterGroup) {
	r.POST("/config-groups", controller.CreateGroup)
	r.PUT("/config-groups/:group", controller.UpdateGroup)
	r.DELETE("/config-groups/:group", controller.DeleteGroup)
	r.POST("/config-items", controller.SaveDefinition)
	r.PUT("/config-items/:id", controller.SaveDefinition)
	r.DELETE("/config-items/:id", controller.DeleteDefinition)
	r.GET("/configs", controller.List)
	r.PUT("/configs/:group", controller.Save)
}
