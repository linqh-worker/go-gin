package rbac

import (
	"strconv"

	auth "go-gin/app/system/logic/auth"
	logic "go-gin/app/system/logic/rbac"
	request "go-gin/app/system/request/rbac"
	"go-gin/internal/errorx"
	"go-gin/internal/httpx"
	"go-gin/internal/httpx/validation"
)

func id(ctx *httpx.Context) (int, error) {
	n, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || n < 1 {
		return 0, errorx.New(10001, "ID必须为正整数")
	}
	return n, nil
}
func Roles(ctx *httpx.Context) (any, error) {
	rows, err := logic.Roles(ctx.Request.Context())
	return rows, auth.HTTPError(err)
}
func AdminRoles(ctx *httpx.Context) (any, error) {
	n, err := id(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := logic.AdminRoles(ctx.Request.Context(), n)
	return rows, auth.HTTPError(err)
}
func SaveRole(ctx *httpx.Context) (any, error) {
	n := 0
	var err error
	if ctx.Param("id") != "" {
		n, err = id(ctx)
		if err != nil {
			return nil, err
		}
	}
	var r request.RoleReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.SaveRole(ctx.Request.Context(), n, r))
}
func SetAdminRoles(ctx *httpx.Context) (any, error) {
	n, err := id(ctx)
	if err != nil {
		return nil, err
	}
	var r request.AdminRolesReq
	if err := ctx.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.SetAdminRoles(ctx.Request.Context(), n, r.RoleIDs))
}
func Admins(ctx *httpx.Context) (any, error) {
	rows, err := logic.Admins(ctx.Request.Context())
	return rows, auth.HTTPError(err)
}
