package auth

import (
	"encoding/json"
	"errors"
	"go-gin/const/errcode"
	"go-gin/internal/errorx"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// validationError 按具体字段转换绑定错误，不向客户端暴露参数原值。
func validationError(err error, request any) error {
	// 项目验证器把字段错误翻译为 BizError；用同一引擎取回字段信息。
	var translated errorx.BizError
	if errors.As(err, &translated) && translated.Code == errorx.ErrCodeValidateFailed {
		if engine, ok := binding.Validator.Engine().(*validator.Validate); ok {
			err = engine.Struct(request)
		}
	}
	var fields validator.ValidationErrors
	if errors.As(err, &fields) && len(fields) > 0 {
		field := fields[0]
		switch field.StructField() {
		case "KeyID", "Timestamp", "Nonce":
			return errorx.New(20008, "缺少或无效的 encryption 加密参数，请重新获取公钥并加密")
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
				return errorx.New(20008, "请提交加密后的确认密码")
			}
			return errorx.New(20008, "确认密码密文格式不正确")
		case "Password":
			if field.Tag() == "required" {
				return errorx.New(20008, "请提交加密后的密码")
			}
			return errorx.New(20008, "密码密文格式不正确")
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
		case "encryption", "encryption.key_id", "encryption.timestamp", "encryption.nonce":
			return errorx.New(20008, "encryption 加密参数类型不正确")
		case "captcha_id", "captcha_code":
			return errorx.New(20006, "captcha_id 和 captcha_code 必须是字符串")
		case "username":
			return errcode.NewDefault("用户名必须是字符串")
		case "password", "password_confirmation":
			return errorx.New(20008, "密码和确认密码密文必须是字符串")
		case "nickname":
			return errcode.NewDefault("昵称必须是字符串")
		}
	}
	return errcode.NewDefault("请求体必须是格式正确的 JSON 对象")
}
