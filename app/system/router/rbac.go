package router

import (
	controller "go-gin/app/system/controller/rbac"
	"go-gin/internal/httpx"
)

func RegisterRBACRoutes(r *httpx.RouterGroup) {
	r.GET("/admins/:id/roles", controller.AdminRoles)
	r.GET("/roles", controller.Roles)
	r.POST("/roles", controller.SaveRole)
	r.PUT("/roles/:id", controller.SaveRole)
	r.PUT("/admins/:id/roles", controller.SetAdminRoles)
}
