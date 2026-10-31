package auth

type CaptchaReq struct {
	Purpose string `form:"purpose" binding:"required,oneof=login" label:"验证码用途"`
}
