package auth

type RegisterReq struct {
	ClientIP             string `json:"-" form:"-"`
	Nickname             string `json:"nickname" form:"nickname" binding:"max=50" label:"昵称"`
	Username             string `json:"username" form:"username" binding:"required,min=3,max=32" label:"用户名"`
	Password             string `json:"password" form:"password" binding:"required,max=72" label:"密码"`
	PasswordConfirmation string `json:"password_confirmation" form:"password_confirmation" binding:"required,max=72" label:"确认密码"`
	CaptchaID            string `json:"captcha_id" form:"captcha_id" binding:"required,max=64" label:"验证码ID"`
	CaptchaCode          string `json:"captcha_code" form:"captcha_code" binding:"required,len=5,numeric" label:"验证码"`
}
