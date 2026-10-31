package user

import (
	logic "go-gin/app/api/logic/auth"
	userLogic "go-gin/app/api/logic/user"
	"go-gin/app/api/middleware"
	"go-gin/internal/httpx"
)

type userController struct {
}

var UserController = &userController{}

func (c *userController) Me(ctx *httpx.Context) (any, error) {
	user, ok := ctx.Get(middleware.CurrentUserKey)
	if !ok {
		return nil, logic.HTTPError(logic.ErrUnauthorized)
	}
	ctx.Header("Cache-Control", "no-store")
	return user, nil
}

func (c *userController) LoginOut(ctx *httpx.Context) (any, error) {
	token := ctx.GetString(middleware.CurrentTokenKey)
	return nil, logic.HTTPError(userLogic.NewUserService().Logout(ctx.Request.Context(), token))
}
