package user

import (
	"context"
	authLogic "go-gin/app/api/logic/auth"
)

// UserService 处理用户登录后的业务。
type UserService struct{}

func NewUserService() *UserService { return &UserService{} }

func (s *UserService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return authLogic.ErrUnauthorized
	}
	return authLogic.NewAuthService().Store.Revoke(ctx, token)
}
