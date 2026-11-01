// 数据库迁移工具
// 用法: go run cmd/migrate/main.go -f .env
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	filex "go-gin/internal/file"
	"go-gin/internal/migration"

	"github.com/labstack/gommon/color"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// 导入迁移文件
	_ "go-gin/migration"
)

var configFile = flag.String("f", "./.env", "the config file")

type DBConfig struct {
	DSN string `yaml:"dsn"`
}

type Config struct {
	DB DBConfig `yaml:"db"`
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	color.Enable()

	var c Config
	if err := filex.MustLoad(*configFile, &c); err != nil {
		return errors.New("读取迁移配置失败，请检查配置文件")
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                      c.DB.DSN,
		DisableDatetimePrecision: true,
		DontSupportRenameIndex:   true,
		DontSupportRenameColumn:  true,
	}), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		return errors.New("数据库连接失败，请检查迁移配置与数据库状态")
	}

	migration.SetDB(db)

	if err := migration.GetManager().Run(); err != nil {
		return errors.New("迁移失败，请检查上方日志及迁移记录")
	}
	return nil
}
