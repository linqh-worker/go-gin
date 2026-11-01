package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	authResponse "go-gin/app/system/response/auth"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

const captchaTTL = 2 * time.Minute
const tokenTTL = 7 * 24 * time.Hour

var ErrCaptcha = errors.New("验证码错误、已过期或已使用")
var ErrUnauthorized = errors.New("登录已失效，请重新登录")

func digest(value string) string {
	h := sha256.Sum256([]byte(value))
	return hex.EncodeToString(h[:])
}

type AuthStore struct{ Redis *redis.Client }

func (s AuthStore) ready() error {
	if s.Redis == nil {
		return errors.New("Redis 未初始化")
	}
	return nil
}

func (s AuthStore) Captcha(ctx context.Context, purpose string) (*authResponse.CaptchaResp, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	if purpose != "login" {
		return nil, errors.New("后台验证码用途必须为 login")
	}
	// 答案使用 crypto/rand 生成，图片驱动只负责绘制。
	buf := make([]byte, 5)
	for i := range buf {
		var b [1]byte
		for {
			if _, err := rand.Read(b[:]); err != nil {
				return nil, err
			}
			if b[0] < 250 {
				break
			}
		}
		buf[i] = '0' + b[0]%10
	}
	answer := string(buf)
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	item, err := driver.DrawCaptcha(answer)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	err = s.Redis.Set(ctx, "system:auth:captcha:"+purpose+":"+id, digest(id+":"+answer), captchaTTL).Err()
	if err != nil {
		return nil, err
	}
	return &authResponse.CaptchaResp{CaptchaID: id, Image: item.EncodeB64string(), ExpiresIn: int(captchaTTL.Seconds())}, nil
}

// GET+DEL 在 Redis 内原子执行，错误答案也消耗该验证码，避免并发重放。
var consumeCaptcha = redis.NewScript(`local v = redis.call('GET', KEYS[1]); if not v then return '' end; redis.call('DEL', KEYS[1]); return v`)

func (s AuthStore) VerifyCaptcha(ctx context.Context, purpose, id, code string) error {
	if err := s.ready(); err != nil {
		return err
	}
	if id == "" || code == "" || len(id) > 64 || len(code) != 5 {
		return ErrCaptcha
	}
	value, err := consumeCaptcha.Run(ctx, s.Redis, []string{"system:auth:captcha:" + purpose + ":" + id}).Text()
	if err != nil {
		return err
	}
	if value == "" || subtle.ConstantTimeCompare([]byte(value), []byte(digest(id+":"+code))) != 1 {
		return ErrCaptcha
	}
	return nil
}

func (s AuthStore) IssueToken(ctx context.Context, userID int, passwordHash string) (string, error) {
	if err := s.ready(); err != nil {
		return "", err
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)
	// 单条 SET 同时写入用户ID与过期时间，不产生永久有效的半成品登录态。
	if err := s.Redis.Set(ctx, "system:auth:token:"+digest(token), strconv.Itoa(userID)+":"+digest(passwordHash), tokenTTL).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (s AuthStore) UserID(ctx context.Context, token string) (int, string, error) {
	if err := s.ready(); err != nil {
		return 0, "", err
	}
	if len(token) != 64 {
		return 0, "", ErrUnauthorized
	}
	// 使用 EXISTS 后 GET 会产生竞态；Lua 直接返回值或空串。
	value, err := s.Redis.Eval(ctx, `return redis.call('GET', KEYS[1]) or ''`, []string{"system:auth:token:" + digest(token)}).Text()
	if err != nil {
		return 0, "", err
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return 0, "", ErrUnauthorized
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		return 0, "", ErrUnauthorized
	}
	return id, parts[1], nil
}

func (s AuthStore) Revoke(ctx context.Context, token string) error {
	if err := s.ready(); err != nil {
		return err
	}
	return s.Redis.Del(ctx, "system:auth:token:"+digest(token)).Err()
}
