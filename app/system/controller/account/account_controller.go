package account

import (
	logic "go-gin/app/system/logic/account"
	auth "go-gin/app/system/logic/auth"

	request "go-gin/app/system/request/account"
	"go-gin/internal/errorx"
	"go-gin/internal/httpx"
	"go-gin/internal/httpx/validation"
	"strconv"
)

func id(c *httpx.Context) (int, error) {
	n, err := strconv.Atoi(c.Param("id"))
	if err != nil || n < 1 {
		return 0, errorx.New(10001, "ID必须为正整数")
	}
	return n, nil
}
func List(admin bool) httpx.HandlerFunc {
	return func(c *httpx.Context) (any, error) {
		var r request.ListReq
		if err := c.ShouldBindQuery(&r); err != nil {
			return nil, validation.Error(err, &r)
		}
		data, err := logic.List(c.Request.Context(), admin, r)
		return data, auth.HTTPError(err)
	}
}
func Detail(admin bool) httpx.HandlerFunc {
	return func(c *httpx.Context) (any, error) {
		n, err := id(c)
		if err != nil {
			return nil, err
		}
		data, err := logic.Detail(c.Request.Context(), admin, n)
		return data, auth.HTTPError(err)
	}
}
func Create(admin bool) httpx.HandlerFunc {
	return func(c *httpx.Context) (any, error) {
		var r request.CreateReq
		if err := c.ShouldBindJSON(&r); err != nil {
			return nil, validation.Error(err, &r)
		}
		data, err := logic.Create(c.Request.Context(), admin, r)
		return data, auth.HTTPError(err)
	}
}
func Update(admin bool) httpx.HandlerFunc {
	return func(c *httpx.Context) (any, error) {
		n, err := id(c)
		if err != nil {
			return nil, err
		}
		var r request.UpdateReq
		if err := c.ShouldBindJSON(&r); err != nil {
			return nil, validation.Error(err, &r)
		}
		return nil, auth.HTTPError(logic.Update(c.Request.Context(), admin, n, r))
	}
}
func Status(admin bool) httpx.HandlerFunc {
	return func(c *httpx.Context) (any, error) {
		n, err := id(c)
		if err != nil {
			return nil, err
		}
		var r request.StatusReq
		if err := c.ShouldBindJSON(&r); err != nil {
			return nil, validation.Error(err, &r)
		}
		return nil, auth.HTTPError(logic.Status(c.Request.Context(), admin, n, c.GetInt("user_id"), *r.Status))
	}
}

func Password(c *httpx.Context) (any, error) {
	n, err := id(c)
	if err != nil {
		return nil, err
	}
	var r request.PasswordReq
	if err := c.ShouldBindJSON(&r); err != nil {
		return nil, validation.Error(err, &r)
	}
	return nil, auth.HTTPError(logic.Password(c.Request.Context(), n, r))
}
