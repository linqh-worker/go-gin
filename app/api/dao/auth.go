package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	modeldao "go-gin/model/dao"
	"go-gin/model/do"
	"gorm.io/gorm"
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
	n, err := modeldao.User.Ctx(ctx).WherePri(id).Where(do.User{Status: 1}).WhereNull("deleted_at").Data(map[string]any{"last_login_at": at, "last_login_ip": ip, "login_count": gorm.Expr("login_count + ?", 1), "updated_at": at}).Update()
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
	err := modeldao.User.Ctx(ctx).Fields(authColumns).Where(do.User{Name: name}).WhereNull("deleted_at").One(&user)
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
	err := modeldao.User.Ctx(ctx).Fields(authColumns).WherePri(id).WhereNull("deleted_at").One(&user)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, nil
	}
	return &user, nil
}

func (AuthRepository) Create(ctx context.Context, user *AuthUser) error {
	id, err := modeldao.User.Ctx(ctx).Data(authData(user)).InsertAndGetId()
	if err == nil {
		user.ID = int(id)
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrUsernameExists
	}
	return err
}

func authData(u *AuthUser) do.User {
	return do.User{Name: u.Name, Nickname: u.Nickname, AvatarUrl: u.AvatarURL, Phone: u.Phone, Email: u.Email, Gender: u.Gender, RegisterIp: u.RegisterIP, LastLoginAt: u.LastLoginAt, LastLoginIp: u.LastLoginIP, LoginCount: u.LoginCount, PasswordChangedAt: u.PasswordChangedAt, PasswordHash: u.PasswordHash, Age: u.Age, Status: u.Status, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt, DeletedAt: u.DeletedAt}
}
