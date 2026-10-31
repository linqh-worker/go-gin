package auth

import (
	"errors"
	"go-gin/app/system/dao"
	"go-gin/const/errcode"
	"go-gin/internal/errorx"
	"net/http"
)

// HTTPError 隐藏基础设施错误细节。
func HTTPError(err error) error {
	if err == nil {
		return nil
	}
	var biz errorx.BizError
	if errors.As(err, &biz) {
		return biz
	}
	switch {
	case errors.Is(err, ErrUnauthorized):
		return errorx.NewServerError(http.StatusUnauthorized)
	case errors.Is(err, ErrCredentials):
		return errcode.ErrUserNameOrPwdFaild
	case errors.Is(err, ErrDisabled):
		return errorx.New(20005, "账号已停用")
	case errors.Is(err, ErrCaptcha):
		return errorx.New(20006, ErrCaptcha.Error())
	case errors.Is(err, dao.ErrUsernameExists):
		return errorx.New(20007, "用户名已存在")
	default:
		return errorx.NewServerError(http.StatusInternalServerError)
	}
}
