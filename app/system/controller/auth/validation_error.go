package auth

import (
	"encoding/json"
	"errors"
	"go-gin/const/errcode"
	"go-gin/internal/errorx"

	"github.com/go-playground/validator/v10"
	"go-gin/internal/httpx/validation"
)

// validationError 按具体字段转换绑定错误，不向客户端暴露参数原值。
func validationError(err error, request any) error {
	err = validation.Normalize(err, request)
	var fields validator.ValidationErrors
	if errors.As(err, &fields) && len(fields) > 0 {
		field := fields[0]
		switch field.StructField() {
		case "CaptchaID":
			if field.Tag() == "required" {
				return errorx.New(20006, "请先获取验证码，并提交 captcha_id")
			}
			return errorx.New(20006, "验证码ID格式不正确，请重新获取验证码")
		case "CaptchaCode":
			if field.Tag() == "required" {
				return errorx.New(20006, "请输入验证码")
			}
			return errorx.New(20006, "验证码必须是5位数字")
		case "PasswordConfirmation":
			if field.Tag() == "required" {
				return errcode.NewDefault("请输入确认密码")
			}
			return errcode.NewDefault("确认密码过长，UTF-8编码不能超过72字节")
		case "Password":
			if field.Tag() == "required" {
				return errcode.NewDefault("请输入密码")
			}
			return errcode.NewDefault("密码过长，UTF-8编码不能超过72字节")
		case "Username":
			if field.Tag() == "required" {
				return errcode.NewDefault("请输入用户名")
			}
			return errcode.NewDefault("用户名须为3至32位字母、数字或下划线")
		case "Nickname":
			return errcode.NewDefault("昵称最多50个字符")
		}
	}
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) {
		switch typeError.Field {
		case "captcha_id", "captcha_code":
			return errorx.New(20006, "captcha_id 和 captcha_code 必须是字符串")
		case "username":
			return errcode.NewDefault("用户名必须是字符串")
		case "password", "password_confirmation":
			return errcode.NewDefault("密码和确认密码必须是字符串")
		case "nickname":
			return errcode.NewDefault("昵称必须是字符串")
		}
	}
	return validation.Error(err, nil)
}
