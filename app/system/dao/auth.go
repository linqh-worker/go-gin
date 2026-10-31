package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"go-gin/internal/component/db"
	"time"
)

var ErrUsernameExists = errors.New("username already exists")

// AuthUser 是用户端认证的数据对象，不作为 HTTP 响应返回。
// 放在模块内，重新生成公共 entity 时不会覆盖认证代码。
type AuthUser struct {
	ID                int        `gorm:"column:id;primaryKey"`
	Name              string     `gorm:"column:name"`
	Nickname          string     `gorm:"column:nickname"`
	AvatarURL         string     `gorm:"column:avatar_url"`
	Phone             *string    `gorm:"column:phone"`
	Email             *string    `gorm:"column:email"`
	Gender            uint8      `gorm:"column:gender"`
	RegisterIP        string     `gorm:"column:register_ip"`
	LastLoginAt       *time.Time `gorm:"column:last_login_at"`
	LastLoginIP       string     `gorm:"column:last_login_ip"`
	LoginCount        int64      `gorm:"column:login_count"`
	PasswordChangedAt *time.Time `gorm:"column:password_changed_at" json:"-"`
	PasswordHash      string     `gorm:"column:password_hash" json:"-"`
	Age               *int       `gorm:"column:age"`
	Status            int        `gorm:"column:status"`
	CreatedAt         *time.Time `gorm:"column:created_at"`
	UpdatedAt         *time.Time `gorm:"column:updated_at"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
}

type AuthRepository struct{}

const authColumns = "id, name, nickname, avatar_url, phone, email, gender, register_ip, last_login_at, last_login_ip, login_count, password_changed_at, password_hash, age, status, created_at, updated_at, deleted_at"

// RecordLogin 使用数据库原子自增，避免并发登录丢失计数。
func (AuthRepository) RecordLogin(ctx context.Context, id int, ip string, at time.Time) error {
	result, err := db.Exec(ctx, "UPDATE `system_admin` SET last_login_at=?, last_login_ip=?, login_count=login_count+1, updated_at=? WHERE id=? AND status=1 AND deleted_at IS NULL", at, ip, at, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrAccountUnavailable
	}
	return nil
}

var ErrAccountUnavailable = errors.New("account unavailable")

func (AuthRepository) FindByName(ctx context.Context, name string) (*AuthUser, error) {
	var user AuthUser
	err := db.GetScan(ctx, &user, "SELECT "+authColumns+" FROM `system_admin` WHERE name = ? AND deleted_at IS NULL LIMIT 1", name)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, nil
	}
	return &user, nil
}

func (AuthRepository) FindByID(ctx context.Context, id int) (*AuthUser, error) {
	var user AuthUser
	err := db.GetScan(ctx, &user, "SELECT "+authColumns+" FROM `system_admin` WHERE id = ? AND deleted_at IS NULL LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, nil
	}
	return &user, nil
}

func (AuthRepository) Create(ctx context.Context, user *AuthUser) error {
	_, err := db.Insert(ctx, "system_admin", user)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrUsernameExists
	}
	return err
}
