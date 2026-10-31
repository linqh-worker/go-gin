package auth

import (
	logic "go-gin/app/system/logic/auth"
	authRequest "go-gin/app/system/request/auth"
	"go-gin/const/errcode"
	"go-gin/internal/httpx"
)

type loginController struct{}

var LoginController = &loginController{}

func (c *loginController) Captcha(ctx *httpx.Context) (any, error) {
	var req authRequest.CaptchaReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, errcode.NewDefault("purpose 必须为 login")
	}
	resp, err := logic.NewAuthService().Store.Captcha(ctx.Request.Context(), req.Purpose)
	ctx.Header("Cache-Control", "no-store")
	return resp, logic.HTTPError(err)
}

func (c *loginController) Login(ctx *httpx.Context) (any, error) {
	var req authRequest.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, validationError(err, &req)
	}
	req.ClientIP = ctx.RemoteIP()
	service := logic.NewAuthService()
	resp, err := service.Login(ctx.Request.Context(), req)
	ctx.Header("Cache-Control", "no-store")
	return resp, logic.HTTPError(err)
}
