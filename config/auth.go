package config

import "go-gin/internal/passwordcrypto"

type Auth struct {
	PasswordPrivateKey string `yaml:"password_private_key"`
}

func GetAuthConf() Auth {
	c := instance.Auth
	if c.PasswordPrivateKey == "" {
		c.PasswordPrivateKey = passwordcrypto.DefaultKeyFile
	}
	return c
}
