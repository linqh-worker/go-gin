// 代码生成工具集
// 用法:
//
//	go run ./cmd/make make:enum                    生成枚举代码
//	go run ./cmd/make make:dao                     生成所有表的 dao/entity/do
//	go run ./cmd/make make:dao -t user,order       生成指定表的 dao/entity/do
//	go run ./cmd/make make:migration <name>        生成迁移文件
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "make:auth-key":
		runAuthKey(args)
	case "make:auth-key-backup":
		runAuthKeyBackup(args)
	case "make:auth-key-restore":
		runAuthKeyRestore(args)
	case "make:enum":
		runEnum()
	case "make:dao":
		runDao(args)
	case "make:migration":
		runMigration(args)
	default:
		fmt.Printf("未知命令: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`代码生成工具集

用法:
  go run ./cmd/make <command> [options]

命令:
  make:auth-key                持久化密码传输RSA公私钥，保留已有文件
  make:auth-key-backup         按key_id版本归档公钥和加密私钥，可提交Git
  make:auth-key-restore -i ... 从加密备份恢复公私钥，不覆盖已有文件
  make:enum                    扫描 const/enum 目录，生成枚举方法
  make:dao                     扫描数据库，生成 entity/do/dao 代码
  make:dao -t user,order       生成指定表的代码
  make:migration <name>        生成迁移文件

示例:
  go run ./cmd/make make:enum
  go run ./cmd/make make:dao -f .env
  go run ./cmd/make make:dao -f .env -t user,order
  go run ./cmd/make make:migration create_orders`)
}
