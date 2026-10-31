// 创建后台管理员，不开放 HTTP 注册入口，也不设置默认密码。
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	admin "go-gin/app/system/dao"
	filex "go-gin/internal/file"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	database, err := gorm.Open(mysql.Open(cfg.DB.DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return errors.New("连接数据库失败，请检查配置")
	}
	pool, err := database.DB()
	if err != nil {
		return err
	}
	defer pool.Close()
	if *grantSuper {
		err := database.Transaction(func(tx *gorm.DB) error {
			var user struct{ ID int }
			if err := tx.Raw("SELECT id FROM system_admin WHERE name=? AND status=1 AND deleted_at IS NULL FOR UPDATE", *name).Scan(&user).Error; err != nil {
				return err
			}
			if user.ID == 0 {
				return errors.New("管理员不存在或已停用")
			}
			var role struct{ ID int }
			if err := tx.Raw("SELECT id FROM system_role WHERE code='super_admin' AND status=1").Scan(&role).Error; err != nil {
				return err
			}
			if role.ID == 0 {
				return errors.New("请先执行 RBAC 迁移")
			}
			return tx.Exec("INSERT INTO system_admin_role(admin_id,role_id) VALUES(?,?) ON DUPLICATE KEY UPDATE role_id=VALUES(role_id)", user.ID, role.ID).Error
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
	if err := database.Table("system_admin").Create(&user).Error; err != nil {
		return errors.New("创建管理员失败，请确认已执行迁移且用户名未被占用；已有账号不会覆盖")
	}
	fmt.Printf("管理员创建成功，ID=%d，用户名=%s\n", user.ID, user.Name)
	return nil
}
