# 后台认证

与用户端共用 HTTP 端口；管理员账号单独保存在 `system_admin`，不查询 `user` 表。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | /system/captcha?purpose=login | 后台登录验证码，5位数字，120秒有效 |
| POST | /system/login | 管理员登录，验证码错误即拒绝 |
| GET | /system/me | 当前管理员，需要后台 Bearer Token |
| POST | /system/logout | 撤销当前后台 Token |

登录请求：

```json
{
  "username": "admin",
  "password": "你的管理员密码",
  "captcha_id": "后台验证码ID",
  "captcha_code": "实际5位数字"
}
```

响应与 API 一致：`code/msg/data/requestId`，成功 code=200；data 包含 token、token_type、expires_in、user。Token 有效期7天，后续通过 `Authorization: Bearer <token>` 传递。用户资料不返回密码哈希；时间使用公共 FormatTime。

## 初始化

```powershell
go run ./cmd/migrate -f ./.env
$adminSecret = Read-Host "管理员密码" -AsSecureString
$env:SYSTEM_ADMIN_PASSWORD = [System.Net.NetworkCredential]::new('', $adminSecret).Password
try {
    go run ./cmd/admin -f ./.env -username admin -nickname 管理员
} finally {
    Remove-Item Env:SYSTEM_ADMIN_PASSWORD
}
```

不开放后台注册，不提供默认账号或默认密码；用户名重复时创建命令不会覆盖原账号。密码至少8字符、UTF-8不超过72字节，存储 bcrypt 哈希。表的字段默认值与用户表一致，但索引及账号数据独立。初始化迁移只创建不存在的表，不修改或清空已有表。

## 隔离规则

- 验证码：`system:auth:captcha:login:<id>`，用户端为 `api:auth:captcha:<purpose>:<id>`。
- Token：`system:auth:token:<摘要>`，用户端为 `api:auth:token:<摘要>`。
- 限流：`system:auth:limit:<action>:<IP摘要>`；验证码每分钟30次，登录10次，不占用用户端配额。
- 请求上下文使用 `system.auth.user` / `system.auth.token`。
- 验证码无论正确或错误均原子消耗；用户端与后台不能互用验证码和 Token。
- 登录中间件每次查询管理员表，账号停用或软删除后拒绝访问。

后台角色与接口权限已实现，使用步骤见 [SYSTEM_RBAC.md](SYSTEM_RBAC.md)。菜单树和数据范围尚未实现；正式部署使用 HTTPS。

个人菜单接口为 GET /system/menus/mine，树形结构和管理接口见 [SYSTEM_MENU.md](SYSTEM_MENU.md)。
