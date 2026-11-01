# 项目使用文档

本篇按当前代码说明配置、命令、接口与运行方式。项目概况见 [README](../README.md)，新增功能见[开发流程说明](DEVELOPMENT.md)，团队规则见[开发约定](AGENTS.md)。除特别注明，命令在项目根目录执行，示例使用PowerShell。

## 目录

- [1. 配置与外部服务](#1-配置与外部服务)
- [2. 启动与命令行](#2-启动与命令行)
- [3. 创建管理员与授予超管](#3-创建管理员与授予超管)
- [4. 迁移和代码生成](#4-迁移和代码生成)
- [5. 用户端接口](#5-用户端接口)
- [6. 后台接口与菜单授权](#6-后台接口与菜单授权)
- [7. 数据表说明](#7-数据表说明)
- [8. 日志与部署](#8-日志与部署)
- [9. 常见问题](#9-常见问题)
- [10. 验证记录](#10-验证记录)

## 1. 配置与外部服务

环境要求：Go1.26.1或更高、MySQL8、Redis（开发约定6.2或更高）。依赖版本以go.mod/go.sum为准。

```powershell
go mod download
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
```

`.env` 是YAML，不能按 `KEY=VALUE` 写。先准备MySQL数据库和账号，迁移不会自动创建数据库。配置完整字段见 [.env.example](../.env.example)，只修改自己的环境文件：

| 配置项 | 含义 |
| --- | --- |
| app.name、app.mode | 应用名；模式debug、test、release，部署使用release |
| app.port | 监听地址，如 `:8088` 或 `127.0.0.1:8088` |
| app.timezone、app.timeformat | 时间配置；用户响应统一经util.FormatTime输出UTC+8 |
| db.dsn | MySQL账号、地址、库名、charset、parseTime和loc |
| db.max-idle-conn、db.max-open-conn | 数据库连接池；分端进程各自持有连接池 |
| db.log-level | silent、debug、info、warn、error |
| redis.addr、redis.username、redis.password、redis.db | 实际Redis实例、认证和逻辑库 |
| log.level、log.path | 日志级别与根目录 |
| monitor.white_ip_list | 队列监控访问IP白名单 |
| svc.user_url、svc.login_url | rest客户端请求的外部系统地址 |
| auth.password_private_key | 历史密码传输密钥配置，当前HTTP认证不使用 |

DSN格式示例仅含占位值：

```text
app_user:change_me@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
```

Redis实例有认证时填写真实用户名/密码，没有认证时留空；NOAUTH说明当前连接的实例要求认证。验证码、登录态、限流和队列共用该Redis配置。队列生产者与消费者必须指向相同实例、逻辑库及队列。

`config.Init()` 每进程只读一次；修改后重启相关进程。文件名如 `.env.testing` 可自定，但必须存在且包含完整YAML，不会因文件名自动改变app.mode。没有通用环境变量覆盖、--env或--conf机制；SYSTEM_ADMIN_PASSWORD等是个别命令专用变量。

## 2. 启动与命令行

CLI由Go标准库flag及os.Args实现，没有Cobra或统一artisan命令。参数可以写 `-f ./.env` 或 `--module=api`；先写make子命令，再写其参数。

### HTTP合并或分端运行

```powershell
# 默认all，使用配置app.port
go run ./cmd/api -f ./.env

# 合并运行在一个端口
go run ./cmd/api -f ./.env --module=all --port=:8088

# 分别在两个终端运行同一程序
go run ./cmd/api -f ./.env --module=api --port=:8088
go run ./cmd/api -f ./.env --module=system --port=:8089
```

module只接受all、api、system；port仅覆盖本次进程的监听地址，不修改文件。两个进程可以共用数据库、Redis及配置，HTTP路径仍是 `/api/...` 和 `/system/...`。每个进程都保留 `/`、`/status`，API模式不挂队列监控；all/system模式挂 `/monitor/queue/*`，受IP白名单保护。

模板端口下：

| 请求 | 用途 |
| --- | --- |
| GET http://127.0.0.1:8088/ | 服务与模块入口信息 |
| GET http://127.0.0.1:8088/api/ | API模块信息，仅all/api |
| GET http://127.0.0.1:8088/status | MySQL、Redis状态；同时检查data字段 |

后台当前没有 `/system/ping`。后台个人信息用 `/system/me`，需要管理员Token。

### 队列和定时任务

```powershell
# 独立终端：消费Asynq任务
go run ./cmd/queue -f ./.env

# 独立终端：调度定时任务
go run ./cmd/cron -f ./.env
```

HTTP只初始化队列投递客户端，不自动运行消费者或调度器。Asynq数据在Redis，不需要MySQL jobs表。cron/init.go仍注册示例任务，使用前调整；同一环境只启动一个调度进程，多个副本会重复执行。

前台进程用Ctrl+C停止；停止HTTP不会停止其他终端进程。

### 命令速查

| 用途 | 命令 | 作用 |
| --- | --- | --- |
| 下载依赖 | `go mod download` | 下载锁定依赖，不整体升级 |
| HTTP帮助 | `go run ./cmd/api -h` | 查看-f、module、port |
| 管理员帮助 | `go run ./cmd/admin -h` | 创建/授予参数 |
| 执行迁移 | `go run ./cmd/migrate -f ./.env` | 执行全部未记录迁移 |
| 新迁移模板 | `go run ./cmd/make make:migration create_orders` | 仅生成模板 |
| 指定表生成 | `go run ./cmd/make make:dao -f ./.env -t user` | 从真实表生成Go代码 |
| 多表生成 | `go run ./cmd/make make:dao -f ./.env -t user,system_menu` | 仅处理指定表 |
| 全库生成 | `go run ./cmd/make make:dao -f ./.env` | 处理库内全部表，使用前审查范围 |
| 枚举生成 | `go run ./cmd/make make:enum` | 扫描const/enum生成方法 |
| 编译检查 | `go build ./...` | 不启动业务或执行迁移 |
| 静态检查 | `go vet ./...` | 不连接数据库 |

queue、cron、migrate也支持-h。`go run ./cmd/make` 不带子命令会打印帮助并返回1。cmd/make包含多个源文件，使用包目录执行，不单独运行main.go。

### Air热重载

```powershell
air -c .air.toml
```

当前配置构建 `./cmd/api` 到 `./tmp/api.exe`，应用参数在 `.air.toml` 的build.args_bin，默认 `["-f", "./.env"]`。Air的-c指定Air配置，应用的-f指定YAML文件。

更换环境可将args_bin改成 `["-f", "./.env.testing", "--module=api", "--port=:8088"]`；需要监听该文件时同步调整include_file，并重启Air。当前配置使用Windows可执行文件名，其他系统按平台调整entrypoint和构建输出。不要与普通go run争用同一端口。

Air需预先安装且位于PATH；安装版本按团队工具约定选定。项目不要求在每次启动前重新安装Air。

## 3. 创建管理员与授予超管

先执行迁移，首次创建账号：

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

用户名3至32位字母数字下划线，转小写；昵称默认用户名、最多50字符。密码至少8字符、UTF-8最多72字节，数据库存bcrypt哈希。账号写入system_admin，不写入用户端user；重复用户名不会覆盖旧账号。

再给已有启用账号授予超管，不需要重新提供密码：

```powershell
go run ./cmd/admin -f ./.env -username admin --grant-super
```

grant-super不修改密码。项目没有固定管理员密码、没有公开后台注册；创建账号与授予超管是两个步骤。普通管理员也可以在后台由超管创建，但新账号没有角色，需要另外分配。

## 4. 迁移和代码生成

```powershell
go run ./cmd/migrate -f ./.env
```

当前migration目录包含六份Create与一份SeedSystem迁移：空库得到六张业务表、migrations记录表、3个角色及27个菜单/接口节点。建表已有则跳过；同名角色不重新授权，菜单按编码或相同方法路径复用，不覆盖既有自定义配置。旧迁移执行记录保留。

迁移按结构体名称末尾14位时间戳排序；已登记的不再执行。部分失败只打印后返回，退出码可能为0，必须检查日志及记录。当前无rollback、status、fresh、reset、step和并发迁移锁，也不自动从数据库导出迁移。

新表开发：

```powershell
go run ./cmd/make make:migration create_orders
# 编辑生成的migration文件，补齐Up、字段、索引、中文注释
go run ./cmd/migrate -f ./.env.testing
go run ./cmd/make make:dao -f ./.env.testing -t orders
gofmt -w model
```

以上testing文件需先准备，orders表需先完成迁移模板。原有表字段变化另建增量迁移，改旧文件不会重跑。

重新生成当前基础表：

```powershell
go run ./cmd/make make:dao -f ./.env -t user,system_admin,system_role,system_menu,system_admin_role,system_role_menu,migrations
gofmt -w model
```

生成entity/do/dao/internal会覆盖；外层dao已存在则跳过。生成器不自动删旧表文件，失败也可能已经写入部分文件；查看diff并编译。生成Go文件提交仓库，克隆后正常编译不应依赖连接数据库先生成。

request/response/controller/logic没有生成命令，需要按开发流程手写。枚举可执行make:enum，但旧enum里的用户状态示例与当前0停用/1启用可能不同，使用前核对，不直接套用。

### 历史密钥维护命令

当前登录注册不使用请求体RSA，也没有 `/api/auth/public-key` 接口。下列工具仅用于维护历史密钥，不是启动步骤：

```powershell
go run ./cmd/make make:auth-key
go run ./cmd/make make:auth-key-backup
go run ./cmd/make make:auth-key-restore -i ./config/keys/实际key_id/private.enc.json
```

make:auth-key支持-o指定私钥路径且不覆盖已有文件；backup支持-i输入私钥、-o归档根目录、-password-file口令文件；restore支持-i备份、-o目标私钥及-password-file，也可用AUTH_KEY_BACKUP_PASSWORD提供口令。默认本地私钥为storage/keys/api_password_private.pem，口令文件与明文私钥不提交；config/keys中的公钥和加密私钥可按开发约定归档，口令需独立备份。

## 5. 用户端接口

### 统一响应与身份

```json
{
  "code": 200,
  "msg": "成功",
  "data": {},
  "requestId": "请求追踪标识"
}
```

外层由config.InitGlobalVars初始化，成功code=200。业务失败通常HTTP仍为200，要同时判断code；401身份无效、403禁止访问、404不存在、429限流，基础设施错误500/503。data.status是账号状态。

| code | 含义 |
| --- | --- |
| 10001 | 通用参数或业务校验失败 |
| 20002 | 用户名或密码错误 |
| 20005 | 账号停用 |
| 20006 | 验证码错误、过期、已使用或专属格式错误 |
| 20007 | 用户名已存在 |

验证码、密码确认等错误分别返回明确提示。多项输入同时错误时，返回先检查到的问题。

后续私有请求携带：

```http
Authorization: Bearer <token>
```

Token为64位十六进制随机字符串，Redis存摘要对应用户ID，有效期604800秒（7天），不是JWT；当前没有refresh-token。多个客户端可分别登录，退出只撤销本次Token。

### 接口清单

| 方法 | 路径 | 身份 |
| --- | --- | --- |
| GET | /api/captcha?purpose=register | 公开 |
| GET | /api/captcha?purpose=login | 公开 |
| POST | /api/register | 公开 |
| POST | /api/login | 公开 |
| GET | /api/me | API Token |
| POST | /api/logout | API Token |

当前没有 `/api/user/` 和demo路由。

### 验证码

```powershell
$base = 'http://127.0.0.1:8088'
$captcha = Invoke-RestMethod "$base/api/captcha?purpose=register"
$captcha.data
```

返回data含captcha_id、image（data:image/png;base64等图片数据）、expires_in=120。前端用image展示，用户填写5位数字；验证码答案不返回，Redis仅保存摘要。

登录、注册按purpose隔离，API与System按端隔离。进入校验步骤后，无论答案正确与否均原子消费；参数绑定失败可能尚未消费，失败后建议重新获取。

### 注册

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
  "captcha_id": "实际注册验证码ID",
  "captcha_code": "实际5位数字"
}
```

nickname可选，最多50字符，空值默认用户名；用户名3至32位字母数字下划线、转小写；密码至少8字符、UTF-8最多72字节，两次一致。captcha_code使用字符串保留前导零。客户端不能指定角色、状态、ID、密码哈希或可信IP。

注册成功自动登录，data与登录相同：

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
    "phone": "",
    "email": "",
    "gender": 0,
    "age": 0,
    "status": 1,
    "last_login_at": "2026-11-01 10:00:00",
    "last_login_ip": "127.0.0.1",
    "login_count": 1,
    "created_at": "2026-11-01 10:00:00",
    "updated_at": "2026-11-01 10:00:00"
  }
}
```

时间为格式示意，统一UTC+8，数据库空时间返回null。不会返回密码哈希、密码设置时间或注册IP。数据库创建与Redis签发不在同一事务，若创建成功但签发/记录登录失败，账号保留，之后通过登录重试，失败不返回Token。

### 登录、当前用户及退出

先取purpose=login验证码，再提交：

```json
{
  "username": "alice_1",
  "password": "ExamplePass123!",
  "captcha_id": "实际登录验证码ID",
  "captcha_code": "实际5位数字"
}
```

POST /api/login成功返回同一登录结构，并记录登录时间、直连IP和次数。GET /api/me的data是user资料；每次重新查账号，停用/软删除后拒绝。POST /api/logout成功data为null，旧Token再次请求返回401。

PowerShell读取个人信息示例：

```powershell
$headers = @{ Authorization = 'Bearer <登录返回的Token>' }
Invoke-RestMethod "$base/api/me" -Headers $headers
Invoke-RestMethod "$base/api/logout" -Method Post -Headers $headers
```

验证码、登录、注册分别按直连IP每分钟30、10、10次限流。返回429及Retry-After；不直接信任X-Forwarded-For，反向代理后的用户可能共用限额。

## 6. 后台接口与菜单授权

### 后台登录

| 方法 | 路径 | 身份要求 |
| --- | --- | --- |
| GET | /system/captcha?purpose=login | 公开，仅支持login用途 |
| POST | /system/login | 公开，字段与API登录相同 |
| GET | /system/me | 后台Token |
| POST | /system/logout | 后台Token |
| GET | /system/menus/mine | 后台Token，可见菜单树 |

后台使用system_admin，响应格式与API相同，管理员资料还包含roles、menu_ids。Token、验证码和限流使用system:auth前缀，API使用api:auth，不能互用；后台验证码每分钟30次、登录10次。密码和账号规则与用户端一致，无公开管理员注册。

### 用户和管理员管理

| 方法 | 用户接口 | 管理员接口 | 用途 |
| --- | --- | --- | --- |
| GET | /system/users | /system/admins | 分页搜索 |
| GET | /system/users/:id | /system/admins/:id | 详情 |
| POST | /system/users | /system/admins | 新增 |
| PATCH | /system/users/:id | /system/admins/:id | 资料修改 |
| PUT | /system/users/:id/status | /system/admins/:id/status | 启停 |

都需要后台Token和相应菜单配置。管理员新增、修改、启停还限制超管；用户写操作可分配给运营角色。

列表参数page默认1、page_size默认20且最多100、keyword搜索用户名/昵称、status为0或1；返回 `items/total/page/page_size`。

创建body为 `{"username":"operator_1","password":"ExamplePass123!","nickname":"运营"}`，不自动登录、不返回Token，新管理员没有角色。修改仅支持nickname、avatar_url、phone、email、age、gender；状态单独 `{"status":0}` 或1。不能改自己管理员状态或通过HTTP停用超管。当前没有账号物理删除与密码重置接口。

### 角色接口

| 方法 | 路径 | 输入/用途 |
| --- | --- | --- |
| GET | /system/roles | 角色列表 |
| POST | /system/roles | 新建：code、name、status |
| PUT | /system/roles/:id | 更新名称、状态 |
| GET | /system/admins/:id/roles | 管理员关联角色ID |
| PUT | /system/admins/:id/roles | 完整替换：`{"role_ids":[实际角色ID]}` |
| GET | /system/roles/:id/menus | 角色关联菜单ID |
| PUT | /system/roles/:id/menus | 完整替换：`{"menu_ids":[实际节点ID]}` |

空数组清空关联；无效/重复ID拒绝并回滚。角色写入、分配与菜单授权仅超管；普通管理员查看列表/管理员角色还需对应接口授权。super_admin为保留角色，只通过服务器命令授予。

默认角色：super_admin超管、auditor只读审核员、user_operator用户运营员。空库seed为只读角色关联用户/管理员/角色读取节点，为运营角色关联用户读取、新增、资料和启停节点；不会自动给任何管理员绑定角色。

### 菜单树与接口权限

关系为 `system_admin → system_admin_role → system_role → system_role_menu → system_menu`。

一个system_menu表保存目录、页面、按钮和API动作，按parent_id关联上下级，根为0，响应children树、叶子children=[]：

| 字段 | 含义 |
| --- | --- |
| code、name | 唯一节点标识、名称；code不是权限编码 |
| parent_id、type | 上级；directory/menu/button/api |
| path、component、icon | 前端页面路径、组件、图标 |
| method、api_path | HTTP方法、后台完整Gin路由模板 |
| sort、status | 排序及0停用/1启用 |

目录下可放目录/页面，页面下放button/api，button/api不能有子节点。最多10层、5000节点，禁止循环；有子节点不能删除。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | /system/menus | 完整树，含停用节点 |
| POST | /system/menus | 新增节点 |
| PUT | /system/menus/:id | 完整更新 |
| DELETE | /system/menus/:id | 删除叶子并清理角色关联 |

管理接口仅超管且需要有效菜单配置。先GET菜单树获取真实页面ID，再新增动作，例如下面parent_id必须换成实际用户列表页面ID：

```json
{
  "parent_id": 2,
  "code": "user_edit_custom",
  "name": "编辑用户",
  "type": "button",
  "method": "PATCH",
  "api_path": "/system/users/:id",
  "sort": 20,
  "status": 1
}
```

按钮/API必须填写method及api_path，不填前端path/component；页面必须填写前端path，可绑定一个接口；目录不绑定接口。一个功能多个接口就在页面下建多个api动作。

前端directory/menu用于导航、button用于按钮、api仅用于授权选择，不渲染侧栏页面。基本目录为“系统管理 → 管理员/角色/菜单管理”和“用户管理 → 用户列表”；后端不会自动生成页面。

### 统一鉴权流程

1. System模块入口Authorize先处理公开接口；其余调用后台TokenCheck。
2. me、menus/mine、logout属于登录即可的个人接口。
3. 其他接口按HTTP方法与 `c.FullPath()` 匹配system_menu.method/api_path，例如PATCH /system/users/12匹配PATCH /system/users/:id。
4. 匹配节点及全部祖先启用；普通管理员通过角色直接关联该节点。超管可跳过角色关联，但不能跳过缺失/停用配置。
5. 授权管理、菜单维护、管理员写入保留超管限制；无配置/无授权/方法不匹配默认403，数据库异常不放行。

给角色勾选节点即可，不维护permission_code、api_id或独立接口规则表。可见树补齐祖先仅用于展示，不授予兄弟或祖先接口；父级复选框需明确提交选中的后代ID。撤权、角色停用和祖先停用立即影响后续请求，无需重新登录。

## 7. 数据表说明

| 表 | 用途 |
| --- | --- |
| user | 用户端账号与资料 |
| system_admin | 后台账号与资料，独立于user |
| system_role | 角色编码、名称及启停 |
| system_admin_role | 管理员与角色关联 |
| system_menu | 目录、页面、按钮、接口及树关系 |
| system_role_menu | 角色与菜单节点关联 |
| migrations | 框架迁移执行记录 |

当前不使用system_api、system_api_rule_legacy、独立permission及role_api等旧授权表，初始化迁移也不创建它们。队列使用Redis，不需要jobs表。不要为清理旧表擅自删除未知业务数据。

user/system_admin主要字段：id主键，name登录名（接口叫username），password_hash bcrypt，nickname、avatar_url、phone、email、age、gender、status，register_ip、last_login_at、last_login_ip、login_count、password_changed_at、created_at、updated_at、deleted_at。

age默认0，phone/email默认空串，status默认1；0停用/1启用，gender为0未知/1男/2女。未知时间可空。name唯一，手机号/邮箱不作为已验证凭据、目前没有唯一索引。新账号显式设置默认资料、创建/更新时间和密码设置时间。

API和System均排除deleted_at不为空的账号。停用只阻断认证，不物理清除Redis Token，重新启用后尚未过期的旧Token可能恢复可用。余额、积分明细、订单等使用单独业务表。

不要求应用表前缀，也不改当前单数表名。迁移和生成DAO显式绑定实际表名，单独改GORM TablePrefix不会自动更新所有代码。

## 8. 日志与部署

日志根目录来自log.path，按access、httpc、access_cron、access_queue等用途分类并按日期保存。模块中间件标记api/system，当前未自动按端分日志文件；按端定位可结合路径、模块和requestId。数据库日志使用参数占位符，业务不输出密码、Token或验证码。

Windows构建：

```powershell
New-Item -ItemType Directory -Path bin -Force
go build -o ./bin/api.exe ./cmd/api
go build -o ./bin/admin.exe ./cmd/admin
go build -o ./bin/migrate.exe ./cmd/migrate
go build -o ./bin/make.exe ./cmd/make
go build -o ./bin/queue.exe ./cmd/queue
go build -o ./bin/cron.exe ./cmd/cron
./bin/api.exe -f ./.env --module=all --port=:8088
```

Linux构建、运行示例：

```bash
mkdir -p bin
go build -o ./bin/api ./cmd/api
./bin/api -f ./.env --module=all --port=:8088
# 或在两个独立进程中分别使用--module=api及--module=system，监听不同端口
```

Windows临时交叉编译Linux amd64，并恢复原环境值：

```powershell
$savedGOOS = $env:GOOS
$savedGOARCH = $env:GOARCH
$savedCGO = $env:CGO_ENABLED
try {
    $env:GOOS = 'linux'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    go build -o ./bin/api ./cmd/api
} finally {
    $env:GOOS = $savedGOOS
    $env:GOARCH = $savedGOARCH
    $env:CGO_ENABLED = $savedCGO
}
```

ARM64机器将GOARCH改为arm64。配置路径相对于工作目录，部署时带上私有配置、固定工作目录，开启HTTPS并限制监控访问。先串行执行迁移、核对结果，再启动业务，worker/cron独立管理。

HTTP目前直接Engine.Run，尚无完整服务超时、自定义优雅退出；迁移无并发锁/回滚，cron无跨副本锁。反向代理部署需明确可信代理：认证IP/限流用直连地址，监控白名单用Gin ClientIP，两者处理不同。

首次上传GitHub：检查.env等私有文件确实被忽略，再审查暂存内容、提交及配置实际远端。已有Git历史或origin时不要重复初始化/覆盖。

```powershell
git check-ignore .env
git status --short --ignored
git add .
git diff --cached --stat
git diff --cached --check
# 审查完整暂存差异后再提交
git commit -m "docs: 统一项目说明、开发流程、约定和使用文档"
# 已有origin则跳过添加，以下地址替换为实际仓库
# git remote add origin https://github.com/你的账号/你的仓库.git
# git push -u origin main
```

## 9. 常见问题

| 现象 | 排查方向 |
| --- | --- |
| MySQL connection refused | MySQL进程、端口、宿主/容器地址 |
| Unknown database | 先创建DSN中的数据库 |
| MySQL Access denied | 实际账号、密码及该库授权 |
| Redis NOAUTH | 当前实例要求认证，核对地址、用户名和密码 |
| 找不到model/do或dao/internal | 是否漏提交生成文件；连开发库重新指定表生成 |
| Unknown column | 连错库或结构未升级；初始化CREATE不会修复已有表 |
| Nothing to migrate | 已登记或没有新迁移；改旧文件不会重跑 |
| HTTP200但业务失败 | 检查code和msg |
| Token无效 | Bearer格式、是否用错端、过期/退出、Redis状态、账号启停 |
| 后台403 | 方法/路由模板、角色菜单关联、节点祖先状态、是否要求超管 |
| 有角色却无菜单 | 新管理员未绑定角色、角色停用、未关联节点；祖先显示不等于授权 |
| 任务不消费 | worker运行、task注册、Redis及队列配置一致 |
| 定时任务重复 | 多个cron进程同时调度 |
| Air找不到main.go | 用项目.air.toml，构建入口为./cmd/api |
| 端口占用 | Air、go run或旧二进制是否同时启动 |
| 缺少encryption参数 | 当前代码已取消要求，检查是否启动了旧二进制 |

## 10. 验证记录

2026-11-01使用原有httpx.Default、路由、中间件、controller、logic和DAO，通过本机真实HTTP完成139项断言。测试使用独立临时MySQL库并执行全部现有迁移，以及与开发配置不同、初始为空的Redis逻辑库。

覆盖注册自动登录、验证码图片/消费/重放/端隔离、密码确认/错误密码/重复账号、bcrypt及主键回填、默认值/时间/登录计数、Token隔离/退出、用户与管理员管理、资料白名单/零值、停用认证拒绝、角色菜单关联、动态路径、未授权方法、超管边界、可见祖先、失败授权保留原状态、角色/祖先停用、清空授权/解绑、菜单树新增/删除/环检查、未配置默认拒绝和限流隔离。

临时环境已删除，原开发库六张业务表内容指纹与验证前一致；临时源文件已移除，不增加永久测试文件，也没有可直接重跑这些断言的现成命令。记录不涵盖浏览器前端、TLS部署、压力/高并发、MySQL或Redis故障恢复；后续代码变化按需重新验证。
