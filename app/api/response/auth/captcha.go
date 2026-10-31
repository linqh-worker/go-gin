package auth

type CaptchaResp struct {
	CaptchaID string `json:"captcha_id"`
	Image     string `json:"image"`
	ExpiresIn int    `json:"expires_in"`
}
