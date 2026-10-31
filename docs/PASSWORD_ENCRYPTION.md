# 密码字段加密与密钥版本管理

登录、注册的 JSON 中，password 和 password_confirmation 必须是 RSA-OAEP-SHA256 密文，不接受明文回退。
只加密密码字段，其他字段保持原格式。后端解密后继续用 bcrypt 保存和校验，已有用户哈希不变，无需数据库迁移。

## 公私钥持久化

```powershell
go run ./cmd/make make:auth-key
```

默认同时保留服务端公私钥，已有私钥不会被重新生成或覆盖：

```text
storage/keys/
├── api_password_private.pem  运行用PKCS8私钥，不提交Git
├── api_password_public.pem   同一私钥派生的公钥
└── backup_password.txt       备份时生成的独立口令，不提交Git
```

启动时读文件，私钥缺失或无效时失败，不会静默换成新密钥。YAML 可配置：

```yaml
auth:
  password_private_key: "storage/keys/api_password_private.pem"
```

省略时使用该默认路径。多个实例应部署同一份当前私钥。

## 加密备份到 Git

```powershell
go run ./cmd/make make:auth-key-backup
```

输出按公钥指纹 key_id 归档的版本目录：

```text
config/keys/<key_id>/
├── public.pem         公钥，可以提交Git
└── private.enc.json   AES-256-GCM加密私钥，可以提交Git
```

备份使用 scrypt 从独立口令派生加密密钥。首次默认生成随机口令，保存在被 Git 忽略的 storage/keys/backup_password.txt；后续复用该口令。
也可通过 AUTH_KEY_BACKUP_PASSWORD 提供至少24字节的自有口令，命令不会打印或自动保存环境变量值。
同版本、同口令重复备份不改写文件，避免无意义的 Git 差异。

**把备份口令另存到密码管理器或独立备份介质。** Git 中只有加密备份，口令也丢失时无法恢复。
提交范围是 config/keys 中的公钥和加密私钥；明文私钥与口令不提交，即使仓库暂时是私有仓库。

## 恢复同一套密钥

先取回原口令文件放到 storage/keys/backup_password.txt，再执行下面命令，将 key_id 替换为实际版本：

```text
go run ./cmd/make make:auth-key-restore -i config/keys/<key_id>/private.enc.json
```

恢复同一份公私钥，不生成新版本、不覆盖已有私钥。可用 -o 指定新的恢复路径，再调整配置并重启。
口令仅保存在密码管理器时，可在 PowerShell 当前进程输入：

```powershell
$backupSecret = Read-Host "原备份口令" -AsSecureString
$env:AUTH_KEY_BACKUP_PASSWORD = [System.Net.NetworkCredential]::new('', $backupSecret).Password
# 执行恢复命令，替换实际版本
Remove-Item Env:AUTH_KEY_BACKUP_PASSWORD
```

命令参数：

```text
make:auth-key -o <私钥路径>
make:auth-key-backup -i <私钥路径> -o <版本备份根目录> -password-file <独立口令文件>
make:auth-key-restore -i <加密备份文件> -o <私钥路径> -password-file <原口令文件>
```

自定义路径时也要确保明文私钥、口令文件被 Git 忽略。

## 前端流程

1. GET /api/auth/public-key 获取 data.key_id、public_key、algorithm。
2. 获取对应 purpose 的图片验证码。
3. 用户提交时加密密码，生成新的时间戳和一次性随机数。
4. 把密文提交到原 /api/register 或 /api/login。

公钥接口 data：

```json
{
  "key_id": "16位公钥指纹",
  "public_key": "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----\n",
  "algorithm": "RSA-OAEP-256",
  "max_age_seconds": 120
}
```

可直接复用无第三方依赖的 [前端模块](../examples/password-encryption.mjs)。后端未托管 examples 静态文件，请将模块复制到前端工程再导入。

```javascript
import { encryptPasswords } from './password-encryption.mjs';

const keys = await fetch('/api/auth/public-key').then(r => r.json());
if (keys.status !== 200) throw new Error(keys.msg);
const body = await encryptPasswords(keys.data, 'register', {
  username: form.username,
  nickname: form.nickname,
  password: form.password,
  password_confirmation: form.passwordConfirmation,
  captcha_id: form.captchaId,
  captcha_code: form.captchaCode
});
const result = await fetch('/api/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(body)
}).then(r => r.json());
```

登录时 purpose 改为 login，不传确认密码。每次重试重新调用 encryptPasswords，不重发旧请求体。
同一个密码每次密文不同，两次密码是否一致在后端解密后判断，不能对密文使用 eqfield。
Web Crypto 需要 HTTPS 或 localhost 安全上下文。Network 会显示密码密文，但表单和前端脚本仍接触原密码。
仍需 HTTPS 认证公钥来源并保护 Token 与其他字段；请求体加密不替代传输层保护。

## 协议与错误

password、password_confirmation 为标准 Base64 RSA 密文，并新增 encryption 对象：

```json
{
  "encryption": {
    "key_id": "公钥接口返回的版本",
    "timestamp": 1793436000,
    "nonce": "32位随机十六进制字符串"
  }
}
```

以上仅为结构示例，不能照抄值。timestamp 使用当前 Unix 秒，nonce 为16字节安全随机数的十六进制编码。
服务端接受过去120秒内的请求，允许最多30秒未来偏差。解密后用 Redis SET NX 原子消耗随机数，之后即使验证码/密码错误也应重新加密再试。
RSA 为2048位，OAEP 与 MGF1 均用 SHA256，UTF-8 label 与前端模块一致：

```text
go-gin:password:v1:<purpose>:<key_id>:<timestamp>:<nonce>:<原始username>:<字段名>
```

字段名是 password 或 password_confirmation，用途是 login 或 register，绑定账号、字段与用途，防止交换密文。

| status | 含义 |
| --- | --- |
| 20008 | 明文、密文损坏、密钥版本或加密参数不正确 |
| 20009 | 请求超时或客户端时间偏差过大 |
| 20010 | 重复提交，请重新加密 |
| 20006 | 解密后验证码错误、过期或已用 |
| 20002 | 验证码通过后用户名或密码错误 |

数据库等内部异常仍映射为500/503，不返回私钥、原密码或解密内部错误。

## 密钥轮换

新私钥生成到新路径，备份后产生新的 key_id 目录。统一更新所有实例的配置并重启。
当前运行时只用一套活跃私钥，不自动加载 Git 里所有历史版本。旧请求在轮换后返回20008，前端重新获取公钥即可。
历史加密备份用于恢复；默认生成命令保留旧密钥，不负责轮换。

实现参考：[Go RSA-OAEP](https://pkg.go.dev/crypto/rsa#EncryptOAEP)、[Web Crypto label](https://developer.mozilla.org/en-US/docs/Web/API/RsaOaepParams)、[Go AEAD](https://pkg.go.dev/crypto/cipher#AEAD)、[scrypt](https://pkg.go.dev/golang.org/x/crypto/scrypt#Key)。
