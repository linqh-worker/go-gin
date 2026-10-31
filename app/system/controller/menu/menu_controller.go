package menu

import (
	auth "go-gin/app/system/logic/auth"
	logic "go-gin/app/system/logic/menu"
	"go-gin/app/system/middleware"
	request "go-gin/app/system/request/menu"
	response "go-gin/app/system/response/user"
	"go-gin/internal/errorx"
	"go-gin/internal/httpx"
	"go-gin/internal/httpx/validation"
	"strconv"
)

func id(ctx *httpx.Context) (int, error) {
	n, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || n < 1 {
		return 0, errorx.New(10001, "ID必须为正整数")
	}
	return n, nil
}
func Tree(ctx *httpx.Context) (any, error) {
	data, err := logic.Tree(ctx.Request.Context())
	return data, auth.HTTPError(err)
}
func Mine(ctx *httpx.Context) (any, error) {
	raw, ok := ctx.Get(middleware.CurrentUserKey)
	if !ok {
		return nil, auth.HTTPError(auth.ErrUnauthorized)
	}
	user, ok := raw.(*response.AuthUser)
	if !ok {
		return nil, auth.HTTPError(auth.ErrUnauthorized)
	}
	ctx.Header("Cache-Control", "no-store")
	data, err := logic.CurrentTree(ctx.Request.Context(), user.ID)
	return data, auth.HTTPError(err)
}
func Save(ctx *httpx.Context) (any, error) {
	n := 0
	var err error
	if ctx.Param("id") != "" {
		n, err = id(ctx)
		if err != nil {
			return nil, err
		}
	}
	var r request.SaveReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.Save(ctx.Request.Context(), n, r))
}
func Delete(ctx *httpx.Context) (any, error) {
	n, err := id(ctx)
	if err != nil {
		return nil, err
	}
	return nil, auth.HTTPError(logic.Delete(ctx.Request.Context(), n))
}
func RoleMenus(ctx *httpx.Context) (any, error) {
	n, err := id(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := logic.RoleMenus(ctx.Request.Context(), n)
	return rows, auth.HTTPError(err)
}
func SetRoleMenus(ctx *httpx.Context) (any, error) {
	n, err := id(ctx)
	if err != nil {
		return nil, err
	}
	var r request.RoleMenusReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.SetRoleMenus(ctx.Request.Context(), n, r.MenuIDs))
}
