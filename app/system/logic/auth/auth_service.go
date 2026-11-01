package auth

import (
	"context"
	"errors"
	"go-gin/app/system/dao"
	authRequest "go-gin/app/system/request/auth"
	authResponse "go-gin/app/system/response/auth"
	userResponse "go-gin/app/system/response/user"
	"go-gin/const/errcode"
	"go-gin/internal/component/redisx"
	"go-gin/util"
	"net"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// UserRepository 让业务可单独测试，实际实现使用模块内 DAO。
type UserRepository interface {
	FindByName(context.Context, string) (*dao.AuthUser, error)
	FindByID(context.Context, int) (*dao.AuthUser, error)
	Create(context.Context, *dao.AuthUser) error
	RecordLogin(context.Context, int, string, time.Time) error
}

type AuthService struct {
	Users UserRepository
	Store AuthStore
}

func NewAuthService() *AuthService {
	return &AuthService{Users: dao.AuthRepository{}, Store: AuthStore{Redis: redisx.Client()}}
}

var usernameRule = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)
var ErrCredentials = errors.New("用户名或密码错误")
var ErrDisabled = errors.New("账号已停用")

func validateCredentials(username, password string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(username))
	if !usernameRule.MatchString(name) {
		return "", errcode.NewDefault("用户名须为3至32位字母、数字或下划线")
	}
	if utf8.RuneCountInString(password) < 8 || len(password) > 72 {
		return "", errcode.NewDefault("密码至少8个字符，且不超过72字节")
	}
	return name, nil
}

func authUser(user *dao.AuthUser) userResponse.AuthUser {
	return userResponse.AuthUser{
		ID:          user.ID,
		Username:    user.Name,
		Nickname:    user.Nickname,
		AvatarURL:   user.AvatarURL,
		Phone:       user.Phone,
		Email:       user.Email,
		Gender:      user.Gender,
		LastLoginAt: util.FormatTime(user.LastLoginAt),
		LastLoginIP: user.LastLoginIP,
		LoginCount:  user.LoginCount,
		UpdatedAt:   util.FormatTime(user.UpdatedAt),
		Age:         user.Age,
		Status:      user.Status,
		CreatedAt:   util.FormatTime(user.CreatedAt),
	}
}

func normalizeIP(value string) string {
	if ip := net.ParseIP(value); ip != nil {
		return ip.String()
	}
	return ""
}

func (s *AuthService) Login(ctx context.Context, req authRequest.LoginReq) (*authResponse.LoginResp, error) {
	name, err := validateCredentials(req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	if err := s.Store.VerifyCaptcha(ctx, "login", req.CaptchaID, req.CaptchaCode); err != nil {
		return nil, err
	}
	user, err := s.Users.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, ErrCredentials
	}
	if user.Status != 1 {
		return nil, ErrDisabled
	}
	return s.createLogin(ctx, user.ID, req.ClientIP)
}

// createLogin 供注册和登录共用，记录登录信息后才返回 Token。
func (s *AuthService) createLogin(ctx context.Context, userID int, ip string) (*authResponse.LoginResp, error) {
	token, err := s.Store.IssueToken(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.Users.RecordLogin(ctx, userID, normalizeIP(ip), time.Now()); err != nil {
		_ = s.Store.Revoke(ctx, token)
		if errors.Is(err, dao.ErrAccountUnavailable) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}
	user, err := s.Users.FindByID(ctx, userID)
	if err != nil || user == nil || user.Status != 1 || user.DeletedAt != nil {
		_ = s.Store.Revoke(ctx, token)
		if err != nil {
			return nil, err
		}
		return nil, ErrUnauthorized
	}
	profile := authUser(user)
	access, err := dao.LoadAccess(ctx, user.ID)
	if err != nil {
		_ = s.Store.Revoke(ctx, token)
		return nil, err
	}
	profile.Roles, profile.MenuIDs = access.Roles, access.MenuIDs
	return &authResponse.LoginResp{Token: token, TokenType: "Bearer", ExpiresIn: int(tokenTTL.Seconds()), User: profile}, nil
}

func (s *AuthService) Current(ctx context.Context, token string) (*userResponse.AuthUser, error) {
	id, err := s.Store.UserID(ctx, token)
	if err != nil {
		return nil, err
	}
	user, err := s.Users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Status != 1 || user.DeletedAt != nil {
		return nil, ErrUnauthorized
	}
	resp := authUser(user)
	access, err := dao.LoadAccess(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	resp.Roles, resp.MenuIDs = access.Roles, access.MenuIDs
	return &resp, nil
}
