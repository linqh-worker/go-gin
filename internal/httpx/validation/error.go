// Package validation 转换请求绑定错误，供不同模块的控制器共用。
package validation

import (
	"encoding/json"
	"errors"
	"go-gin/const/errcode"
	"go-gin/internal/errorx"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Normalize 从项目验证器翻译后的错误中取回字段信息，供业务定制提示。
func Normalize(err error, request any) error {
	var translated errorx.BizError
	if errors.As(err, &translated) && translated.Code == errorx.ErrCodeValidateFailed && request != nil {
		if engine, ok := binding.Validator.Engine().(*validator.Validate); ok {
			if raw := engine.Struct(request); raw != nil {
				return raw
			}
		}
	}
	return err
}

// Error 处理通用格式错误；只返回字段名和规则，不包含请求原值。
func Error(err error, request any) error {
	if err == nil {
		return nil
	}
	err = Normalize(err, request)
	var fields validator.ValidationErrors
	if errors.As(err, &fields) && len(fields) > 0 {
		field := fields[0]
		switch field.Tag() {
		case "required":
			return errcode.NewDefault(field.Field() + "不能为空")
		case "max":
			return errcode.NewDefault(field.Field() + "超过允许的最大值或长度")
		case "min":
			return errcode.NewDefault(field.Field() + "未达到允许的最小值或长度")
		default:
			return errcode.NewDefault(field.Field() + "格式不正确")
		}
	}
	var wrongType *json.UnmarshalTypeError
	if errors.As(err, &wrongType) {
		return errcode.NewDefault(wrongType.Field + "字段类型不正确")
	}
	var translated errorx.BizError
	if errors.As(err, &translated) && translated.Code == errorx.ErrCodeValidateFailed {
		return translated
	}
	return errcode.NewDefault("请求参数格式不正确，请检查请求体及字段类型")
}
