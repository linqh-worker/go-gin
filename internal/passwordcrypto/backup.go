package passwordcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"golang.org/x/crypto/scrypt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultBackupRoot = "config/keys"
const DefaultBackupPasswordFile = "storage/keys/backup_password.txt"
const BackupPasswordEnv = "AUTH_KEY_BACKUP_PASSWORD"

type keyBackup struct {
	Version    int    `json:"version"`
	KeyID      string `json:"key_id"`
	Algorithm  string `json:"algorithm"`
	KDF        string `json:"kdf"`
	Salt       string `json:"salt"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

func writeNewFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	return f.Close()
}

// BackupPassword 优先使用环境口令；只有备份命令允许首次生成本地随机口令。
// 恢复时禁止生成新口令，以免用错误口令误导用户。
func BackupPassword(path string, create bool) (string, error) {
	if value := os.Getenv(BackupPasswordEnv); value != "" {
		if len(value) < 24 {
			return "", errors.New("备份口令至少24字节")
		}
		return value, nil
	}
	data, err := os.ReadFile(path)
	if err == nil {
		value := strings.TrimSpace(string(data))
		if len(value) < 24 {
			return "", errors.New("本地备份口令长度不足")
		}
		return value, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	if !create {
		return "", errors.New("找不到备份口令，请提供 AUTH_KEY_BACKUP_PASSWORD 或原口令文件")
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	value := hex.EncodeToString(buf)
	if err := writeNewFile(path, []byte(value+"\n"), 0600); err != nil {
		return "", err
	}
	return value, nil
}

func backupAEAD(password string, salt []byte) (cipher.AEAD, error) {
	if len(password) < 24 || len(salt) != 16 {
		return nil, errors.New("备份加密参数无效")
	}
	key, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func backupAAD(keyID string) []byte { return []byte("go-gin:password-key-backup:v1:" + keyID) }

// BackupFile 将公钥和加密私钥归档到以 key_id 命名的版本目录；已有版本保持不变。
func BackupFile(privatePath, root, password string) (string, error) {
	manager, err := LoadFile(privatePath)
	if err != nil {
		return "", err
	}
	directory := filepath.Join(root, manager.public.KeyID)
	backupPath := filepath.Join(directory, "private.enc.json")
	if existing, err := os.ReadFile(backupPath); err == nil {
		restored, err := decryptBackup(existing, password)
		if err != nil {
			return "", err
		}
		if restored.public.KeyID != manager.public.KeyID {
			return "", errors.New("已有备份版本与私钥不符")
		}
		// 同版本同口令重复执行时不重写随机密文，避免无意义的 Git 变化。
		return backupPath, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	aead, err := backupAEAD(password, salt)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	der, err := x509.MarshalPKCS8PrivateKey(manager.private)
	if err != nil {
		return "", err
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	ciphertext := aead.Seal(nil, nonce, privatePEM, backupAAD(manager.public.KeyID))
	backup := keyBackup{Version: 1, KeyID: manager.public.KeyID, Algorithm: "AES-256-GCM", KDF: "scrypt-32768-8-1", Salt: base64.StdEncoding.EncodeToString(salt), Nonce: base64.StdEncoding.EncodeToString(nonce), Ciphertext: base64.StdEncoding.EncodeToString(ciphertext)}
	encoded, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return "", err
	}
	publicPath := filepath.Join(directory, "public.pem")
	if existing, err := os.ReadFile(publicPath); err == nil {
		if !bytes.Equal(existing, []byte(manager.public.PublicKey)) {
			return "", errors.New("版本目录已有不同公钥，未覆盖")
		}
	} else if os.IsNotExist(err) {
		if err := writeNewFile(publicPath, []byte(manager.public.PublicKey), 0644); err != nil {
			return "", err
		}
	} else {
		return "", err
	}
	if err := writeNewFile(backupPath, append(encoded, '\n'), 0644); err != nil {
		return "", err
	}
	return backupPath, nil
}

func decryptBackup(data []byte, password string) (*Manager, error) {
	if len(data) > 32<<10 {
		return nil, errors.New("备份文件过大")
	}
	var backup keyBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, errors.New("备份文件格式错误")
	}
	if backup.Version != 1 || backup.Algorithm != "AES-256-GCM" || backup.KDF != "scrypt-32768-8-1" || len(backup.KeyID) != 16 {
		return nil, errors.New("不支持的密钥备份版本")
	}
	salt, e1 := base64.StdEncoding.DecodeString(backup.Salt)
	nonce, e2 := base64.StdEncoding.DecodeString(backup.Nonce)
	ciphertext, e3 := base64.StdEncoding.DecodeString(backup.Ciphertext)
	if e1 != nil || e2 != nil || e3 != nil || len(salt) != 16 || len(nonce) != 12 {
		return nil, errors.New("备份加密参数无效")
	}
	aead, err := backupAEAD(password, salt)
	if err != nil {
		return nil, err
	}
	plain, err := aead.Open(nil, nonce, ciphertext, backupAAD(backup.KeyID))
	if err != nil {
		return nil, errors.New("备份口令错误或文件已损坏")
	}
	manager, err := parsePrivate(plain)
	if err != nil {
		return nil, err
	}
	if manager.public.KeyID != backup.KeyID {
		return nil, errors.New("备份密钥ID校验失败")
	}
	return manager, nil
}

// RestoreFile 恢复同一份私钥，不覆盖已有文件，也不生成新的密钥版本。
func RestoreFile(backupPath, privatePath, password string) error {
	if _, err := os.Stat(privatePath); err == nil {
		return os.ErrExist
	} else if !os.IsNotExist(err) {
		return err
	}
	info, err := os.Stat(backupPath)
	if err != nil {
		return err
	}
	if info.Size() > 32<<10 {
		return errors.New("备份文件过大")
	}
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}
	manager, err := decryptBackup(data, password)
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKCS8PrivateKey(manager.private)
	if err != nil {
		return err
	}
	if err := writeNewFile(privatePath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0600); err != nil {
		return err
	}
	_, err = EnsurePublicFile(privatePath)
	return err
}
