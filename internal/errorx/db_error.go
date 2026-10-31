package errorx

import (
	"errors"

	"gorm.io/gorm"
)

type DBError struct {
	Msg   string
	cause error
}

func (e DBError) Unwrap() error { return e.cause }

func (e DBError) Error() string {
	return e.Msg
}

func TryToDBError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return DBError{Msg: err.Error(), cause: err}
}
