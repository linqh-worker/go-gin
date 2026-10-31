package middleware

import (
	response "go-gin/app/system/response/user"
	"go-gin/internal/errorx"
	"go-gin/internal/httpx"
)

// RequirePermission 放在 TokenCheck 后；没有匹配权限时默认拒绝。
func RequirePermission(code string) httpx.HandlerFunc {
	return func(ctx *httpx.Context) (any, error) {
		raw, ok := ctx.Get(CurrentUserKey)
		if !ok {
			return nil, errorx.NewServerError(401)
		}
		user, ok := raw.(*response.AuthUser)
		if !ok {
			return nil, errorx.NewServerError(401)
		}
		for _, p := range user.Permissions {
			if p == "*" || p == code {
				return nil, nil
			}
		}
		return nil, errorx.NewServerError(403)
	}
}

// RequireSuper 只允许保留角色管理授权，普通角色不能委派权限给自己。
func RequireSuper() httpx.HandlerFunc {
	return func(ctx *httpx.Context) (any, error) {
		raw, ok := ctx.Get(CurrentUserKey)
		if !ok {
			return nil, errorx.NewServerError(401)
		}
		user, ok := raw.(*response.AuthUser)
		if !ok {
			return nil, errorx.NewServerError(401)
		}
		for _, r := range user.Roles {
			if r == "super_admin" {
				return nil, nil
			}
		}
		return nil, errorx.NewServerError(403)
	}
}
