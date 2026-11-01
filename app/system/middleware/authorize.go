package middleware

import (
	"go-gin/app/system/authorization"
	menu "go-gin/app/system/logic/menu"
	response "go-gin/app/system/response/user"
	"go-gin/internal/errorx"
	"go-gin/internal/httpx"
)

func Authorize() httpx.HandlerFunc {
	authenticate := TokenCheck()
	return func(c *httpx.Context) (any, error) {
		method, path := c.Request.Method, c.FullPath()
		if authorization.IsPublic(method, path) {
			return nil, nil
		}
		if _, err := authenticate(c); err != nil {
			return nil, err
		}
		raw, ok := c.Get(CurrentUserKey)
		if !ok {
			return nil, errorx.NewServerError(401)
		}
		user, ok := raw.(*response.AuthUser)
		if !ok {
			return nil, errorx.NewServerError(401)
		}
		if authorization.IsPersonal(method, path) {
			return nil, nil
		}
		super := false
		for _, role := range user.Roles {
			if role == "super_admin" {
				super = true
			}
		}
		if !super && authorization.IsSuperOnly(method, path) {
			return nil, errorx.NewServerError(403)
		}
		allowed, err := menu.CanAccess(c.Request.Context(), method, path, user.MenuIDs, super)
		if err != nil {
			return nil, errorx.NewServerError(503)
		}
		if !allowed {
			return nil, errorx.NewServerError(403)
		}
		return nil, nil
	}
}
