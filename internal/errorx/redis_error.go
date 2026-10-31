package errorx

import (
	"errors"
	"github.com/redis/go-redis/v9"
)

type RedisError struct {
	Msg   string
	cause error
}

func (e RedisError) Unwrap() error { return e.cause }

func TryToRedisError(err error) error {
	if err == nil || errors.Is(err, redis.Nil) {
		return err
	}
	return RedisError{Msg: err.Error(), cause: err}
}

func (e RedisError) Error() string {
	return e.Msg
}
