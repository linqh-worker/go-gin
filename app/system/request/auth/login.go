package auth

type LoginReq struct {
	ClientIP    string `json:"-" form:"-"`
	Username    string `json:"username" form:"username" binding:"required,min=3,max=32" label:"用户名"`
	Password    string `json:"password" form:"password" binding:"required,max=72" label:"密码"`
	CaptchaID   string `json:"captcha_id" form:"captcha_id" binding:"required,max=64" label:"验证码ID"`
	CaptchaCode string `json:"captcha_code" form:"captcha_code" binding:"required,len=5,numeric" label:"验证码"`
}
