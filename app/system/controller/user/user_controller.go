package user

import (
	logic "go-gin/app/system/logic/auth"
	"go-gin/app/system/middleware"
	"go-gin/internal/httpx"
)

type userController struct{}

var UserController = &userController{}

func (*userController) Me(ctx *httpx.Context) (any, error) {
	user, ok := ctx.Get(middleware.CurrentUserKey)
	if !ok {
		return nil, logic.HTTPError(logic.ErrUnauthorized)
	}
	ctx.Header("Cache-Control", "no-store")
	return user, nil
}
func (*userController) Logout(ctx *httpx.Context) (any, error) {
	return nil, logic.HTTPError(logic.NewAuthService().Store.Revoke(ctx.Request.Context(), ctx.GetString(middleware.CurrentTokenKey)))
}
