package main

import (
	"errors"
	"flag"
	"fmt"
	"go-gin/internal/passwordcrypto"
	"os"
)

func runAuthKey(args []string) {
	flags := flag.NewFlagSet("make:auth-key", flag.ExitOnError)
	path := flags.String("o", passwordcrypto.DefaultKeyFile, "密码传输私钥路径；已有文件不会覆盖")
	flags.Parse(args)
	if err := passwordcrypto.GenerateFile(*path); err != nil {
		if errors.Is(err, os.ErrExist) {
			fmt.Println("私钥已存在，保留原文件：", *path)
		} else {
			fmt.Fprintln(os.Stderr, "生成私钥失败：", err)
			os.Exit(1)
		}
	}
	info, err := passwordcrypto.EnsurePublicFile(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "持久化公钥失败：", err)
		os.Exit(1)
	}
	fmt.Println("密码传输私钥：", *path)
	fmt.Println("密码传输公钥：", passwordcrypto.PublicKeyPath(*path))
	fmt.Println("密钥版本 key_id：", info.KeyID)
}

func runAuthKeyBackup(args []string) {
	flags := flag.NewFlagSet("make:auth-key-backup", flag.ExitOnError)
	input := flags.String("i", passwordcrypto.DefaultKeyFile, "本地私钥路径")
	output := flags.String("o", passwordcrypto.DefaultBackupRoot, "可提交Git的加密备份目录")
	passwordFile := flags.String("password-file", passwordcrypto.DefaultBackupPasswordFile, "独立保存的备份口令文件，不提交Git")
	flags.Parse(args)
	password, err := passwordcrypto.BackupPassword(*passwordFile, true)
	if err != nil {
		fmt.Fprintln(os.Stderr, "获取备份口令失败：", err)
		os.Exit(1)
	}
	path, err := passwordcrypto.BackupFile(*input, *output, password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "备份密钥失败：", err)
		os.Exit(1)
	}
	fmt.Println("可提交Git的加密私钥备份：", path)
	if os.Getenv(passwordcrypto.BackupPasswordEnv) != "" {
		fmt.Println("备份口令来自环境变量，请自行独立保管")
	} else {
		fmt.Println("备份口令文件（请另存密码管理器，不提交Git）：", *passwordFile)
	}
}

func runAuthKeyRestore(args []string) {
	flags := flag.NewFlagSet("make:auth-key-restore", flag.ExitOnError)
	input := flags.String("i", "", "加密备份private.enc.json路径（必填）")
	output := flags.String("o", passwordcrypto.DefaultKeyFile, "恢复的本地私钥路径，不覆盖已有文件")
	passwordFile := flags.String("password-file", passwordcrypto.DefaultBackupPasswordFile, "原备份口令文件，也可通过AUTH_KEY_BACKUP_PASSWORD提供")
	flags.Parse(args)
	if *input == "" {
		fmt.Fprintln(os.Stderr, "请用 -i 指定加密备份文件")
		os.Exit(1)
	}
	password, err := passwordcrypto.BackupPassword(*passwordFile, false)
	if err == nil {
		err = passwordcrypto.RestoreFile(*input, *output, password)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "恢复密钥失败，已有文件不会覆盖：", err)
		os.Exit(1)
	}
	fmt.Println("公私钥已恢复：", *output)
}
