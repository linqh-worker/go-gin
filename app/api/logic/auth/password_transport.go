package auth

import (
	"context"
	"go-gin/internal/passwordcrypto"
	"strings"
	"time"
)

// DecryptPasswords 位于业务边界，只把解密后的密码交给现有注册/登录逻辑。
func (s *AuthService) DecryptPasswords(ctx context.Context, meta passwordcrypto.Metadata, purpose, username, password, confirmation string) (string, string, error) {
	manager, err := passwordcrypto.Current()
	if err != nil {
		return "", "", err
	}
	plain, err := manager.Decrypt(meta, purpose, username, "password", password, time.Now())
	if err != nil {
		return "", "", err
	}
	var confirmed string
	if purpose == "register" {
		confirmed, err = manager.Decrypt(meta, purpose, username, "password_confirmation", confirmation, time.Now())
		if err != nil {
			return "", "", err
		}
	}
	if err := s.Store.ready(); err != nil {
		return "", "", err
	}
	// 每次提交的新随机数只允许使用一次，Redis 原子 SET NX 防止多实例并发重放。
	key := "api:auth:encryption_nonce:" + meta.KeyID + ":" + strings.ToLower(meta.Nonce)
	// 时间戳按秒判断，额外保留一秒覆盖 Unix 秒截断的边界。
	fresh, err := s.Store.Redis.SetNX(ctx, key, "1", passwordcrypto.MaxAge+passwordcrypto.FutureSkew+time.Second).Result()
	if err != nil {
		return "", "", err
	}
	if !fresh {
		return "", "", passwordcrypto.ErrReplay
	}
	return plain, confirmed, nil
}
