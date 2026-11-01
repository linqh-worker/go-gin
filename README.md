# go-gin

基于Gin、GORM、MySQL和Redis的Go后端开发基础。采用单应用、多模块结构：用户端 `/api`、管理端 `/system`，可共用一个端口，也可用同一程序启动两个进程分别监听。

已提供用户登录注册、图片验证码、独立后台认证、账号管理、角色与树形菜单授权，以及迁移、DAO生成、队列、定时任务和事件机制。可以在现有分层上继续开发业务；仓库不包含前端页面。

## 文档入口

全项目维护以下四份主文档，根目录只保留本README：

| 文档 | 用途 |
| --- | --- |
| [项目说明](README.md) | 能力、架构、目录和快速开始 |
| [开发流程说明](docs/DEVELOPMENT.md) | 从需求、建表到接口、授权、验证与交付 |
| [开发约定](docs/AGENTS.md) | 人工与AI共同遵守的分层、数据及提交规则 |
| [使用文档](docs/USAGE.md) | 配置、命令、接口调用、后台授权、部署与排错 |

新成员先完成快速开始，再按开发流程实现功能。修改项目之前阅读开发约定；`docs/AGENTS.md` 是普通文档，不会作为根目录规则自动加载，使用AI工具时应明确要求先阅读。

## 环境要求

- Go **1.26.1或更高**，版本以 `go.mod` 为准。
- MySQL **8.0或更高**，当前迁移使用 `utf8mb4_0900_ai_ci`。
- Redis：开发环境约定6.2或更高；认证、验证码、限流和Asynq使用Redis。
- Git；Air仅用于开发热重载，Docker及 `.devcontainer` 为可选环境。

## 快速开始

在项目根目录执行，下列命令使用PowerShell：

```powershell
go version
go mod download
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
```

`.env` 实际是 **YAML文件**。填写数据库、Redis和监听地址，不覆盖已有私有配置。先通过数据库客户端准备数据库及账号，例如：

```sql
CREATE DATABASE demo CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
```

初始化结构并启动：

```powershell
go run ./cmd/migrate -f ./.env
go build ./...
go vet ./...
go run ./cmd/api -f ./.env --module=all
```

迁移创建六张业务表：`user`、`system_admin`、`system_role`、`system_menu`、`system_admin_role`、`system_role_menu`；另有框架的 `migrations` 记录表。空库初始化3个角色和27个菜单节点，不创建默认账号或密码。已有表跳过建表，不自动补齐结构差异。

默认端口为 `:8088`。`GET /` 查看服务信息，`GET /api/` 查看API模块，`GET /status` 查看数据库、Redis状态，同时检查 `data` 中组件状态。首次创建后台管理员见[使用文档](docs/USAGE.md#3-创建管理员与授予超管)。

## 架构与目录

```text
HTTP → router → 本端中间件 → controller → logic → DAO → MySQL
                                ↓
                         response → httpx统一响应
```

`request` 定义输入、`response` 定义输出；业务错误由业务层返回，控制器转换后交给统一响应。`logic` 就是通常说的service层，不额外增加同职责service目录。

```text
app/
├── api/                  用户端：controller、logic、request、response、dao、middleware、router
└── system/               管理端：同上；authorization维护公开/个人/超管接口边界
cmd/                      api、admin、queue、cron、migrate、make命令入口
config/                   YAML配置、响应和环境初始化
const/                    枚举、错误定义
model/                    生成的entity、do、dao
migration/                建表、增量变更和基础数据迁移
internal/                 httpx、数据库、Redis、日志、队列等组件
util/                     公共工具，如时间格式化
rest/                     请求其他系统HTTP接口的客户端
cron/ task/ event/         定时任务、异步任务、事件及注册
router/                   模块选择和公共路由
storage/                  日志、运行文件与本地备份
```

功能按业务拆目录，如 `app/system/controller/account`；model与基础组件可共用，身份及业务权限仍按端隔离。

## 当前能力与边界

| 能力 | 当前实现 |
| --- | --- |
| 用户认证 | 用户名/密码、bcrypt、图片验证码、注册自动登录、Redis随机Token、退出 |
| 后台认证 | 独立管理员表、验证码、Token和限流；无公开注册 |
| 账号管理 | 用户与管理员列表、详情、新增、资料修改、启停 |
| 后台授权 | 管理员关联角色，角色关联树形菜单；菜单保存HTTP方法与接口路径 |
| 数据访问 | 生成的DAO/do/entity、模块DAO、事务与行锁 |
| 迁移和生成 | 时间戳排序Up、执行记录、指定表DAO及枚举生成 |
| 异步与定时 | Asynq独立worker、cron独立进程、事件监听 |
| 日志与追踪 | 分类日志、requestId；尚未按api/system自动分文件 |

Token有效期7天，**不是JWT**。密码直接通过JSON提交，数据库存bcrypt哈希，部署使用HTTPS；当前不要求请求体RSA加密或公钥接口。

目前没有前端页面、密码找回、短信/邮箱验证、用户端自行修改资料、账号物理删除或刷新Token接口；迁移没有回滚命令和并发锁，cron没有跨副本调度锁。HTTP超时及自定义优雅退出需按部署需求完善。

## 验证记录

2026-11-01在独立MySQL库、空闲Redis逻辑库中，通过真实本机HTTP请求完成139项断言，覆盖认证、账号管理、角色菜单授权及主要拒绝路径。临时环境清理完毕，原开发库六表内容指纹一致。详细范围见[使用文档](docs/USAGE.md#10-验证记录)，不代表前端、TLS、并发和故障恢复已验证。

环境私有配置、日志、二进制和数据库备份不提交；源代码、迁移及生成代码随项目提交。后续按[开发流程说明](docs/DEVELOPMENT.md)新增功能。

## 来源与许可证

本项目在已有应用框架基础上二次开发，保留已有源码版权声明。

本仓库二开内容包括：api/system模块划分、独立后台认证、账号资料与管理员改密、基于菜单的角色授权、系统配置分组管理，以及迁移、接口响应和开发文档的调整。具体变更以本仓库Git提交记录为准。

- Gin：[原仓库](https://github.com/gin-gonic/gin)，当前依赖版本v1.9.1，原始MIT许可证见 [licenses/gin/LICENSE](licenses/gin/LICENSE)。
- 已核实的第三方来源与本地适配说明见 [NOTICE](NOTICE)。
- 2026-11-01核对时，该上游未查到明确的许可证文件或授权声明；记录来源不等于获得许可，发布相关上游代码前需进一步核实授权。Gin的MIT许可证不能替代应用框架本身的许可。
