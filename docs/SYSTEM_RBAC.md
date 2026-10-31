# 后台 RBAC 使用与开发规范

权限链路：**管理员 → 多个角色 → 多个权限点**。只作用于 `/system`，不读取用户端身份或角色。

## 初始化与超级管理员

迁移新增四张表，保留原管理员表及数据：

| 表 | 说明 |
| --- | --- |
| system_role | 角色编码、名称、启用状态 |
| system_permission | 代码中实际使用的权限点目录 |
| system_admin_role | 管理员与角色，多对多，联合主键防重复 |
| system_role_permission | 角色与权限，多对多，联合主键防重复 |

```powershell
go run ./cmd/migrate -f ./.env
# admin 必须是已经存在且启用的后台管理员
go run ./cmd/admin -f ./.env -username admin --grant-super
```

创建账号见 [后台认证](SYSTEM_AUTH.md)。授予超级管理员命令不会修改密码，也不会自动创建账号。现有账号不会自动变成超级管理员。

保留角色 `super_admin` 具备所有权限，返回 `permissions: ["*"]`。不能通过 HTTP 创建、修改、分配或移除这个角色；普通角色管理与角色分配也仅允许超级管理员操作，避免授权接口成为提权入口。

## 接口与权限

除登录、验证码外，以下接口都携带后台 `Authorization: Bearer <token>`。

| 方法 | 路径 | 需要权限 |
| --- | --- | --- |
| GET | /system/me | 已登录，返回本人角色与权限 |
| GET | /system/admins | system.admin.read，最多返回前100个未删除账号的基本资料 |
| GET | /system/admins/:id/roles | system.admin.read |
| GET | /system/roles | system.role.read |
| GET | /system/roles/:id/permissions | system.role.read，返回权限ID数组 |
| GET | /system/permissions | system.permission.read |
| POST | /system/roles | 仅超级管理员 |
| PUT | /system/roles/:id | 仅超级管理员 |
| PUT | /system/admins/:id/roles | 仅超级管理员 |

登录响应的 `data.user` 和 `/system/me` 的 `data` 都包含：

```json
{
  "roles": ["operator"],
  "permissions": ["system.role.read"]
}
```

没有角色时返回两个空数组，可以查看本人资料和退出，访问受限接口返回 HTTP403。未登录、账号停用或软删除返回 HTTP401。

## 创建角色并分配

1. 用超级管理员 Token 请求 `GET /system/permissions`，取得实际权限ID。
2. 请求 `POST /system/roles`：

```json
{
  "code": "operator",
  "name": "运营人员",
  "status": 1,
  "permission_ids": [1, 2]
}
```

示例ID需要替换为目录中实际存在的值。角色编码为3至64位小写字母、数字或下划线，以字母开头，创建后不修改。

3. 请求 `GET /system/roles` 取得角色ID，再请求 `PUT /system/admins/管理员ID/roles`：

```json
{"role_ids": [2]}
```

`role_ids` 是完整替换而非追加。传 `[]` 可清空普通管理员角色；关联变更使用事务，不存在或重复的ID会拒绝并回滚。

更新角色使用 `PUT /system/roles/角色ID`：

```json
{"name":"运营人员","status":0,"permission_ids":[]}
```

`status` 为0停用、1启用；`permission_ids` 为完整替换，可传空数组撤销权限。保留角色不允许通过接口修改。当前以停用角色代替删除，不提供角色删除接口。

## 添加业务权限

权限点是服务端定义的操作，不按客户端传入的路由或按钮名称自动授权。新增接口时：

1. 新建增量迁移，向 `system_permission` 插入稳定编码，例如 `system.order.read`。
2. 把路由放到 `TokenCheck()` 后，再通过专用子组的 `Before()` 添加权限检查：

```go
orders := private.Group("")
orders.Before(middleware.RequirePermission("system.order.read"))
orders.GET("/orders", controller.List)
```

3. 超级管理员把该权限分配给角色。普通管理员即获得对应接口访问权限。

本项目 httpx 的普通路由 handler 会立即写响应，所以权限函数必须使用 `Before()`，不要和最终控制器并列作为 GET/POST 的多个 handler。

每次认证都会重新读取数据库里的启用角色与权限，撤权、停用不依赖 Token 过期；已经在执行的请求不保证被追溯取消。角色权限控制操作能力，业务数据归属、订单可见范围等仍由业务层校验。

菜单树与按钮节点已实现，见 [菜单树文档](SYSTEM_MENU.md)；当前未实现数据范围、授权审计和权限缓存。前端可以根据权限隐藏按钮，服务端仍必须执行权限中间件。
