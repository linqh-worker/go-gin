package authorization

func IsPublic(method, path string) bool {
	return method == "GET" && path == "/system/captcha" || method == "POST" && path == "/system/login"
}
func IsPersonal(method, path string) bool {
	return method == "GET" && (path == "/system/me" || path == "/system/menus/mine") || method == "POST" && path == "/system/logout"
}
func IsSuperOnly(method, path string) bool {
	switch method + " " + path {
	case "POST /system/config-groups", "PUT /system/config-groups/:group", "DELETE /system/config-groups/:group",
		"POST /system/config-items", "PUT /system/config-items/:id", "DELETE /system/config-items/:id",
		"PUT /system/admins/:id/password", "POST /system/admins", "PATCH /system/admins/:id", "PUT /system/admins/:id/status", "PUT /system/admins/:id/roles",
		"POST /system/roles", "PUT /system/roles/:id", "GET /system/roles/:id/menus", "PUT /system/roles/:id/menus",
		"GET /system/menus", "POST /system/menus", "PUT /system/menus/:id", "DELETE /system/menus/:id":
		return true
	}
	return false
}
