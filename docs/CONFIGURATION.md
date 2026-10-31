# 配置与部署

## 配置来源

配置模板是根目录 [.env.example](../.env.example)。本项目使用 YAML 解析；不要按 dotenv 或 Viper 项目的 KEY=VALUE 方式填写。
各入口通过 `-f` 读取一个配置文件，没有通用环境变量覆盖机制，也没有因 `.env.testing` 文件名自动改变模式的机制。
`config.Init()` 在进程内只读取一次，改配置后重启相关进程，不支持热加载。

PowerShell 首次复制：

```powershell
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
```

Linux/macOS 可使用 `cp -n .env.example .env`。所有相对路径以启动时的工作目录为准。

## 字段

| 配置 | 作用 |
| --- | --- |
| app.name | 应用标识 |
| app.port | HTTP 监听地址，例如 :8088，注意保留冒号 |
| app.mode | debug、test、release，影响 Gin 模式和错误展示 |
| app.timezone / timeformat | 时间处理配置，默认模板使用 Asia/Shanghai |
| auth.password_private_key | 密码传输私钥文件，省略为 storage/keys/api_password_private.pem |
| db.dsn | MySQL 账号、地址、库名、字符集、parseTime 和时区 |
| db.max-idle-conn / max-open-conn | 数据库连接池上限，按实际连接预算设置 |
| db.log-level | silent、debug、info、warn、error |
| redis.addr | Redis 地址及端口 |
| redis.username / password | Redis 实际认证信息，没有认证时为空字符串 |
| redis.db | 当前应用和队列使用的逻辑库编号 |
| monitor.white_ip_list | /monitor/queue/* 的客户端 IP 白名单 |
| log.level | debug、info、warn、error |
| log.path | 日志根目录，末尾可以带 / |
| svc.user_url / login_url | rest 客户端示例的外部服务地址 |

应用缓存、验证码、登录态、限流和队列当前共用 redis 配置；不能把 Redis 逻辑库当作租户或权限隔离措施。
HTTP、queue、cron 应使用匹配的配置，否则可能出现投递和消费不在同一 Redis 的问题。

## 环境区分

可分别维护 `.env`、`.env.testing`、`.env.production`，每个文件填写完整 YAML：

```powershell
go run ./cmd/api -f ./.env.testing
go run ./cmd/queue -f ./.env.testing
```

这些文件都不提交仓库。示例只保存可公开的占位值，不放真实密码、访问令牌或业务数据。
生产配置使用 `app.mode: release`、适当的日志级别，限制队列监控白名单。
当前仓库不负责自动申请数据库账号、发布配置、创建数据库或部署 Redis。

## 日志与追踪

当前日志根目录取 `log.path`，按用途写入 access、httpc、access_cron、access_queue 等子目录，默认日志按日期命名。
响应和请求追踪字段由 `config/var.go` 的 `requestId` 定义。开发业务日志传递 context，便于关联同一请求。
模块中间件会标记 api/system，但当前没有自动按端分文件的实现。
数据库日志保留参数占位符，Redis 失败日志只记录命令名称；新增日志也不得输出密码、完整 Token、验证码答案或整份配置。

## 部署与进程

1. 在目标 Go/OS/架构环境构建二进制，准备配置和外部服务。
2. 对目标数据库执行迁移，检查输出与迁移记录，成功后再启动新版本业务。
3. HTTP、worker 和 cron 使用独立进程管理，分别观察日志和重启。
4. 只启动业务需要的进程；cron 的示例任务在 `cron/init.go` 中，需要先按需求调整。

目前 HTTP 直接使用 Engine.Run，没有配置完整的服务超时及自定义优雅退出流程；cron 没有跨副本调度锁；迁移没有并发锁、回滚命令或可靠的失败退出码。
因此部署流程要串行执行迁移，cron 保持单调度实例。需要多实例或更严格运行保障时，先补齐对应机制，不能仅靠增加副本解决。

## 认证和代理

API 的注册/登录 IP 和限流按直连地址取值，不直接信任 X-Forwarded-For；代理后面的用户目前可能共用限额。
队列监控使用 Gin ClientIP 和白名单。项目尚未建立统一可信代理策略，部署反向代理时必须同时复核这些地址处理方式，不能把监控路由直接对公网开放。
System 只有公开 ping。后台增加敏感接口前要实现自己的管理员身份及权限，不能共用用户端 Token 验证。

## 常见问题

| 现象 | 先检查什么 |
| --- | --- |
| MySQL connection refused | 服务是否运行、端口是否正确、是否误连宿主/容器地址 |
| Unknown database | DSN 对应的数据库是否已创建 |
| Access denied | 数据库用户名、密码和该库授权 |
| Redis NOAUTH | 当前连接的实例要求认证，核对实际地址和保存的用户名/密码 |
| 找不到 model/do 或 dao/internal | 是否漏提交生成文件；准备数据库后按指定表重新生成 |
| Unknown column | 配置连接是否正确，表是否与迁移一致；建表快照不会修复旧表 |
| 迁移 Nothing to migrate | 已登记执行或没有新迁移；改旧文件不会自动重跑 |
| HTTP 200 但业务失败 | 检查外层 code 与 msg，不只看 HTTP 状态码 |
| Token 无效 | Bearer 格式、是否退出/过期、Redis 登录态、用户是否停用或软删除 |
| 任务不消费 | worker 是否运行、task/init.go 是否注册、两边 Redis 与队列名是否一致 |
| 定时任务重复 | 是否同时启动多个 cron 进程 |
| Air 找不到 main.go | 使用本项目 .air.toml，构建入口应为 ./cmd/api，而不是根目录 main.go |
