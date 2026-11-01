// 创建后台管理员，不开放 HTTP 注册入口，也不设置默认密码。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go-gin/internal/component/db"
	modeldao "go-gin/model/dao"
	"go-gin/model/do"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	admin "go-gin/app/system/dao"
	filex "go-gin/internal/file"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	path := flag.String("f", "./.env", "配置文件")
	name := flag.String("username", "", "管理员用户名")
	nickname := flag.String("nickname", "", "昵称，默认用户名")
	grantSuper := flag.Bool("grant-super", false, "向已有管理员授予超级管理员角色，不修改密码")
	flag.Parse()
	*name = strings.ToLower(strings.TrimSpace(*name))
	if !regexp.MustCompile(`^[a-z0-9_]{3,32}$`).MatchString(*name) {
		return errors.New("用户名须为3至32位字母、数字或下划线")
	}
	password := os.Getenv("SYSTEM_ADMIN_PASSWORD")
	if !*grantSuper && (utf8.RuneCountInString(password) < 8 || len(password) > 72) {
		return errors.New("请通过 SYSTEM_ADMIN_PASSWORD 环境变量提供至少8字符、不超过72字节的密码")
	}
	*nickname = strings.TrimSpace(*nickname)
	if *nickname == "" {
		*nickname = *name
	}
	if utf8.RuneCountInString(*nickname) > 50 {
		return errors.New("昵称最多50个字符")
	}
	var cfg struct {
		DB struct {
			DSN string `yaml:"dsn"`
		} `yaml:"db"`
	}
	filex.MustLoad(*path, &cfg)
	db.InitConfig(db.Config{DSN: cfg.DB.DSN, MaxOpenConns: 2, MaxIdleConns: 1, LogLevel: "silent"})
	if err := db.Connect(); err != nil {
		return errors.New("连接数据库失败，请检查配置")
	}
	ctx := context.Background()
	if *grantSuper {
		err := db.Transaction(ctx, func(tx *db.TX) error {
			var user struct{ ID int }
			if err := modeldao.SystemAdmin.Ctx(ctx).WithTX(tx).Where(do.SystemAdmin{Name: *name, Status: 1}).WhereNull("deleted_at").LockUpdate().One(&user); err != nil {
				return err
			}
			if user.ID == 0 {
				return errors.New("管理员不存在或已停用")
			}
			var role struct{ ID int }
			if err := modeldao.SystemRole.Ctx(ctx).WithTX(tx).Where(do.SystemRole{Code: "super_admin", Status: 1}).One(&role); err != nil {
				return err
			}
			if role.ID == 0 {
				return errors.New("请先执行 RBAC 迁移")
			}
			q := modeldao.SystemAdminRole.Ctx(ctx).WithTX(tx)
			exists, err := q.Where(do.SystemAdminRole{AdminId: user.ID, RoleId: role.ID}).Exist()
			if err != nil || exists {
				return err
			}
			_, err = q.Data(do.SystemAdminRole{AdminId: user.ID, RoleId: role.ID}).Insert()
			return err
		})
		if err != nil {
			return errors.New("授予超级管理员失败，请确认账号有效且已执行 RBAC 迁移")
		}
		fmt.Println("已授予超级管理员角色：", *name)
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now()
	empty := ""
	age := 0
	user := admin.AuthUser{Name: *name, Nickname: *nickname, PasswordHash: string(hash), Status: 1, Phone: &empty, Email: &empty, Age: &age, PasswordChangedAt: &now, CreatedAt: &now, UpdatedAt: &now}
	if err := (admin.AuthRepository{}).Create(ctx, &user); err != nil {
		return errors.New("创建管理员失败，请确认已执行迁移且用户名未被占用；已有账号不会覆盖")
	}
	fmt.Printf("管理员创建成功，ID=%d，用户名=%s\n", user.ID, user.Name)
	return nil
}
