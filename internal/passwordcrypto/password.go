// Package passwordcrypto 实现登录、注册密码字段的 RSA-OAEP-SHA256 传输加密。
package passwordcrypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

const DefaultKeyFile = "storage/keys/api_password_private.pem"
const MaxAge = 120 * time.Second
const FutureSkew = 30 * time.Second

var ErrInvalid = errors.New("密码密文无效，请重新获取公钥并加密")
var ErrExpired = errors.New("加密请求已过期，请重新加密后提交")
var ErrReplay = errors.New("请勿重复提交，请重新加密后重试")
var active atomic.Pointer[Manager]

type Metadata struct {
	KeyID     string `json:"key_id" binding:"required,len=16,hexadecimal" label:"公钥ID"`
	Timestamp int64  `json:"timestamp" binding:"required,gt=0" label:"加密时间戳"`
	Nonce     string `json:"nonce" binding:"required,len=32,hexadecimal" label:"加密随机数"`
}

type PublicInfo struct {
	KeyID         string `json:"key_id"`
	PublicKey     string `json:"public_key"`
	Algorithm     string `json:"algorithm"`
	MaxAgeSeconds int64  `json:"max_age_seconds"`
}

type Manager struct {
	private *rsa.PrivateKey
	public  PublicInfo
}

func New(private *rsa.PrivateKey) (*Manager, error) {
	if private == nil || private.N == nil || private.N.BitLen() != 2048 {
		return nil, errors.New("密码传输密钥须为 RSA-2048")
	}
	if err := private.Validate(); err != nil {
		return nil, err
	}
	der, err := x509.MarshalPKIXPublicKey(&private.PublicKey)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(der)
	return &Manager{private: private, public: PublicInfo{
		KeyID: hex.EncodeToString(hash[:8]), PublicKey: string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})),
		Algorithm: "RSA-OAEP-256", MaxAgeSeconds: int64(MaxAge.Seconds()),
	}}, nil
}

func Init(path string) error {
	manager, err := LoadFile(path)
	if err != nil {
		return err
	}
	active.Store(manager)
	return nil
}

func LoadFile(path string) (*Manager, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parsePrivate(data)
}

func parsePrivate(data []byte) (*Manager, error) {
	block, rest := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" || len(strings.TrimSpace(string(rest))) != 0 {
		return nil, errors.New("私钥必须是 PKCS8 PEM PRIVATE KEY")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("无法解析密码传输私钥")
	}
	private, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("密码传输私钥不是 RSA")
	}
	return New(private)
}

func Current() (*Manager, error) {
	manager := active.Load()
	if manager == nil {
		return nil, errors.New("密码传输私钥尚未初始化")
	}
	return manager, nil
}

func (m *Manager) Public() PublicInfo { return m.public }

// Label 绑定用途、账号、字段、时间戳和随机数，不能把注册密文移到登录或其他账号。
func Label(meta Metadata, purpose, username, field string) []byte {
	return []byte(strings.Join([]string{"go-gin", "password", "v1", purpose, meta.KeyID, strconv.FormatInt(meta.Timestamp, 10), meta.Nonce, username, field}, ":"))
}

func (m *Manager) Decrypt(meta Metadata, purpose, username, field, ciphertext string, now time.Time) (string, error) {
	if meta.KeyID != m.public.KeyID || (purpose != "login" && purpose != "register") || (field != "password" && field != "password_confirmation") {
		return "", ErrInvalid
	}
	if len(meta.Nonce) != 32 {
		return "", ErrInvalid
	}
	if _, err := hex.DecodeString(meta.Nonce); err != nil {
		return "", ErrInvalid
	}
	if meta.Timestamp < now.Add(-MaxAge).Unix() || meta.Timestamp > now.Add(FutureSkew).Unix() {
		return "", ErrExpired
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(ciphertext)
	if err != nil || len(decoded) != m.private.Size() {
		return "", ErrInvalid
	}
	plain, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, m.private, decoded, Label(meta, purpose, username, field))
	if err != nil || !utf8.Valid(plain) {
		return "", ErrInvalid
	}
	return string(plain), nil
}

// GenerateFile 不覆盖已有私钥；多实例应共享部署的同一份私钥。
func GenerateFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return os.ErrExist
	} else if !os.IsNotExist(err) {
		return err
	}
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKCS8PrivateKey(private)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if err := pem.Encode(f, &pem.Block{Type: "PRIVATE KEY", Bytes: der}); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	_, err = EnsurePublicFile(path)
	return err
}

func PublicKeyPath(privatePath string) string {
	if strings.HasSuffix(privatePath, "_private.pem") {
		return strings.TrimSuffix(privatePath, "_private.pem") + "_public.pem"
	}
	return strings.TrimSuffix(privatePath, filepath.Ext(privatePath)) + "_public.pem"
}

// EnsurePublicFile 从已有私钥派生并持久化公钥，不生成或替换私钥。
func EnsurePublicFile(privatePath string) (PublicInfo, error) {
	manager, err := LoadFile(privatePath)
	if err != nil {
		return PublicInfo{}, err
	}
	path := PublicKeyPath(privatePath)
	if existing, err := os.ReadFile(path); err == nil {
		block, _ := pem.Decode(existing)
		expected, _ := pem.Decode([]byte(manager.public.PublicKey))
		if block == nil || block.Type != "PUBLIC KEY" || !bytes.Equal(block.Bytes, expected.Bytes) {
			return PublicInfo{}, errors.New("已有公钥与私钥不匹配，文件未覆盖")
		}
		return manager.Public(), nil
	} else if !os.IsNotExist(err) {
		return PublicInfo{}, err
	}
	if err := writeNewFile(path, []byte(manager.public.PublicKey), 0644); err != nil {
		return PublicInfo{}, err
	}
	return manager.Public(), nil
}
