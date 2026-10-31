package auth

import (
	logic "go-gin/app/api/logic/auth"
	authRequest "go-gin/app/api/request/auth"
	"go-gin/const/errcode"
	"go-gin/internal/httpx"
	"go-gin/internal/passwordcrypto"
)

type loginController struct{}

var LoginController = &loginController{}

func (c *loginController) PublicKey(ctx *httpx.Context) (any, error) {
	ctx.Header("Cache-Control", "no-store")
	manager, err := passwordcrypto.Current()
	if err != nil {
		return nil, logic.HTTPError(err)
	}
	return manager.Public(), nil
}

func (c *loginController) Captcha(ctx *httpx.Context) (any, error) {
	var req authRequest.CaptchaReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, errcode.NewDefault("purpose 必须为 login 或 register")
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
	plain, _, err := service.DecryptPasswords(ctx.Request.Context(), req.Encryption, "login", req.Username, req.Password, "")
	if err != nil {
		return nil, logic.HTTPError(err)
	}
	req.Password = plain
	resp, err := service.Login(ctx.Request.Context(), req)
	ctx.Header("Cache-Control", "no-store")
	return resp, logic.HTTPError(err)
}

func (c *loginController) Register(ctx *httpx.Context) (any, error) {
	var req authRequest.RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, validationError(err, &req)
	}
	req.ClientIP = ctx.RemoteIP()
	service := logic.NewAuthService()
	plain, confirmation, err := service.DecryptPasswords(ctx.Request.Context(), req.Encryption, "register", req.Username, req.Password, req.PasswordConfirmation)
	if err != nil {
		return nil, logic.HTTPError(err)
	}
	req.Password, req.PasswordConfirmation = plain, confirmation
	resp, err := service.Register(ctx.Request.Context(), req)
	return resp, logic.HTTPError(err)
}
