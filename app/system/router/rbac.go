package router

import (
	controller "go-gin/app/system/controller/rbac"
	"go-gin/app/system/middleware"
	"go-gin/internal/httpx"
)

func RegisterRBACRoutes(r *httpx.RouterGroup) {
	private := r.Group("")
	private.Before(middleware.TokenCheck())
	admins := private.Group("")
	admins.Before(middleware.RequirePermission("system.admin.read"))
	admins.GET("/admins", controller.Admins)
	admins.GET("/admins/:id/roles", controller.AdminRoles)
	roles := private.Group("")
	roles.Before(middleware.RequirePermission("system.role.read"))
	roles.GET("/roles", controller.Roles)
	roles.GET("/roles/:id/permissions", controller.RolePermissions)
	permissions := private.Group("")
	permissions.Before(middleware.RequirePermission("system.permission.read"))
	permissions.GET("/permissions", controller.Permissions)
	manage := private.Group("")
	manage.Before(middleware.RequireSuper())
	manage.POST("/roles", controller.SaveRole)
	manage.PUT("/roles/:id", controller.SaveRole)
	manage.PUT("/admins/:id/roles", controller.SetAdminRoles)
}
