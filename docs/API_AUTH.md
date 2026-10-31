# API 认证接口

默认示例地址为 `http://127.0.0.1:8088`，实际端口以配置为准。认证使用用户名、密码和图片验证码；不包含短信、邮件或后台管理员认证。
先按 [README](../README.md) 准备数据库、Redis 并启动 HTTP。

密码直接通过 JSON 字符串提交，不需要公钥或 encryption 参数。数据库继续使用 bcrypt 哈希；部署使用 HTTPS。

## 接口表

| 方法 | 路径 | 登录要求 |
| --- | --- | --- |
| GET | /api/captcha?purpose=register | 无，注册图片验证码 |
| GET | /api/captcha?purpose=login | 无，登录图片验证码 |
| POST | /api/register | 无，创建用户 |
| POST | /api/login | 无，校验密码并创建登录态 |
| GET | /api/me | Bearer Token |
| GET | /api/user/ | Bearer Token，与 me 相同 |
| POST | /api/logout | Bearer Token，撤销当前登录态 |

登录、注册使用 JSON 请求体。登录和退出不接受 GET。登录身份传入 `Authorization: Bearer <token>`，不使用旧的 token 请求头。

## 响应约定

正式入口的统一响应来自 `config/var.go`：

```json
{
  "code": 200,
  "msg": "成功",
  "data": {},
  "requestId": "追踪标识"
}
```

外层 code 是业务码，成功为200。业务失败通常 HTTP 状态仍为200，需要判断外层 code；身份失效返回 HTTP 401，限流429，基础服务异常500或503。
直接在测试里创建 httpx.New() 不会自动调用入口的初始化；正式接口测试要显式调用 config.InitGlobalVars() 才能验证相同字段。

| 外层 code | 含义 |
| --- | --- |
| 200 | 成功 |
| 10001 | 当前认证控制器的通用参数/业务校验错误 |
| 20002 | 用户名或密码错误 |
| 20005 | 账号已停用 |
| 20006 | 验证码错误、过期或已使用 |
| 20007 | 用户名已存在 |
| 401 | 需要重新登录 |
| 429 | 请求次数超过限额 |

参数绑定不再统一提示“请检查注册参数、两次密码和5位验证码”：验证码缺失、格式错误或答案校验失败使用20006，并明确提示对应字段；确认密码不一致单独提示。JSON 格式错误、密码长度等仍使用10001。
业务顺序为参数绑定、密码规则及验证码、账号密码校验；多项错误时返回先检查到的问题。

## 获取验证码

```http
GET /api/captcha?purpose=register
```

data 字段示例：

```json
{
  "captcha_id": "随机ID",
  "image": "data:image/png;base64,...",
  "expires_in": 120
}
```

前端用 image 作为 img.src 展示，把用户看到的5位数字与 captcha_id 一起提交。验证码答案不返回给客户端；Redis 保存摘要。
注册与登录验证码按 purpose 隔离，有效期两分钟，进入验证码验证步骤后正确或错误答案均会消耗它。参数绑定失败可能尚未进入该步骤；调用失败后重新获取验证码最明确。
Redis 内原子读取并删除，不能并发重复使用同一验证码。

## 注册

```http
POST /api/register
Content-Type: application/json
```

```json
{
  "username": "alice_1",
  "nickname": "小明",
  "password": "ExamplePass123!",
  "password_confirmation": "ExamplePass123!",
  "captcha_id": "注册验证码ID",
  "captcha_code": "12345"
}
```

上面的12345仅示意格式，必须替换为实际图片数字。

- username：3至32位字母、数字或下划线，业务层转为小写并去掉首尾空白。
- nickname：可选，最多50字符；空值默认用户名，不要求唯一。
- password：直接提交密码，至少8个字符、UTF-8编码不超过72字节。注册时两次密码必须一致。
- captcha_code：字符串格式的5位数字，保留前导零。

密码以 bcrypt 保存；新用户状态固定为1。客户端无法指定权限、账号状态、用户ID、密码哈希和记录IP。
注册成功后自动登录，返回与登录相同的 token、token_type、expires_in 和 user，并记录首次登录时间、IP和次数。用户名唯一性由数据库唯一索引约束，并发重复注册也返回业务错误。
ClientIP 是服务端内部参数，JSON 不接受该字段。

## 登录

重新获取 `purpose=login` 的验证码，然后提交：

```http
POST /api/login
Content-Type: application/json
```

```json
{
  "username": "alice_1",
  "password": "ExamplePass123!",
  "captcha_id": "登录验证码ID",
  "captcha_code": "12345"
}
```

注册与登录成功响应 data 的结构相同，如下，user 展示完整字段；时间仅为格式示例：

```json
{
  "token": "随机Token",
  "token_type": "Bearer",
  "expires_in": 604800,
  "user": {
    "id": 1,
    "username": "alice_1",
    "nickname": "小明",
    "avatar_url": "",
    "phone": null,
    "email": null,
    "gender": 0,
    "last_login_at": "2026-10-31 10:00:00",
    "last_login_ip": "127.0.0.1",
    "login_count": 1,
    "updated_at": "2026-10-31 10:00:00",
    "age": null,
    "status": 1,
    "created_at": "2026-10-31 09:00:00"
  }
}
```

data.user.status 是账号状态，不是外层业务码。密码哈希、密码设置时间和注册IP不在用户响应中。
登录通过后记录时间、直连IP和次数；记录写入失败会尝试撤销新 Token，客户端不会获得该 Token。

当前 Token 是加密安全随机字符串，登录态保存在 Redis，有效期7天，**不是 JWT**，没有 refresh-token 接口。
Redis key 使用 Token 摘要。多个客户端可分别登录；退出只撤销当前 Token。Redis 登录态丢失后需要重新登录。

## 当前用户与退出

```http
GET /api/me
Authorization: Bearer 随机Token
```

me 的 data 与登录返回的 user 字段相同。身份中间件重新读取账号，停用、软删除、Token 过期或已撤销都会拒绝访问。

```http
POST /api/logout
Authorization: Bearer 随机Token
```

退出成功返回统一外层响应，data 为 null；同一 Token 再请求 me 返回401。过期 Token 请求退出也会被身份中间件拒绝。

## 限流与已知边界

验证码、登录、注册分别按直连IP限制为每分钟30、10、10次。超限返回429和 Retry-After；代理部署时多个用户可能共用限额，需先建立可信代理策略再调整。

当前尚无用户资料编辑、密码找回、手机/邮箱验证接口；这些资料字段存在不代表已经验证。
旧用户如果 password_hash 为空，不能通过密码登录；未实现自动初始化密码。
原全表用户列表、添加和批量添加路由不再默认开放。`/api/demo/*` 是登录后的示例，不是后台管理授权。
后台认证见 [SYSTEM_AUTH.md](SYSTEM_AUTH.md)，用户端 Token 不承担管理员身份。

## 代码入口

| 位置 | 职责 |
| --- | --- |
| app/api/router/login.go | 公开认证路由和限流 |
| app/api/router/user.go | 用户路由与 TokenCheck |
| app/api/controller/auth | 参数解析、认证业务调用 |
| app/api/controller/user | 当前用户和退出 |
| app/api/logic/auth | 验证码、密码和登录态业务 |
| app/api/logic/user | 登录后用户业务 |
| app/api/request/auth | 输入结构和标签 |
| app/api/response/auth | 登录与验证码响应 |
| app/api/response/user | 用户资料响应 |
| app/api/dao/auth.go | 用户端认证数据库访问 |

表结构见 [USER_SCHEMA.md](USER_SCHEMA.md)，测试命令见 [COMMANDS.md](COMMANDS.md)。

用户响应中的 last_login_at、created_at、updated_at 统一为北京时间（UTC+8）的 yyyy-MM-dd HH:mm:ss，数据库空时间返回 null。

注册创建账号与 Redis 登录态不属于同一事务。若账号创建成功但签发 Token 或记录登录失败，接口返回错误，账号保留，可通过登录接口重试；失败时不返回 Token。
