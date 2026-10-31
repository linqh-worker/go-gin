# go-gin

基于 Gin 的 Go 后端项目，采用单应用、单 HTTP 端口、多业务模块结构。用户端使用 `/api`，管理端预留 `/system`，按功能拆分控制器、业务、请求和响应。

项目提供可继续扩展的开发基础，不包含完整后台管理系统或前端界面。

## 文档导航

| 文档 | 内容 |
| --- | --- |
| [命令行说明](docs/COMMANDS.md) | 启动、迁移、代码生成、测试、编译 |
| [模块与架构](docs/MODULES.md) | 目录职责、加载顺序、模块边界 |
| [开发规范](docs/DEVELOPMENT.md) | 统一分层、命名、接口、数据库和测试规则，含新增业务示例 |
| [配置与部署](docs/CONFIGURATION.md) | YAML 配置、日志、运行进程、故障排查 |
| [认证接口](docs/API_AUTH.md) | 注册、登录、图片验证码、当前用户、退出 |
| [密码加密与密钥备份](docs/PASSWORD_ENCRYPTION.md) | 前端密文、公私钥持久化、版本备份与恢复 |
| [用户表](docs/USER_SCHEMA.md) | 字段、索引、初始化迁移与表名前缀 |
| [协作与提交](docs/CONTRIBUTING.md) | 分支、提交、GitHub 首次上传及变更交付 |
| [AI 开发约定](docs/AGENTS.md) | 后续自动化修改必须遵守的项目规则 |

新成员按「本页快速开始 → 模块与架构 → 开发规范 → 对应接口文档」阅读。

所有详细说明集中在 docs/。后续人工或 AI 开发前先阅读开发规范与 AI 开发约定；AGENTS.md、CLAUDE.md 已作为文档收纳，不再位于根目录供工具自动加载。

## 环境要求

- Go **1.26.1 或更高**，以 `go.mod` 为准。
- MySQL **8.0 或更高**，当前建表快照使用 `utf8mb4_0900_ai_ci`。
- Redis：团队开发环境约定使用6.2或更高版本；用户认证、验证码、限流与 Asynq 队列使用同一配置实例。
- Git；Docker 仅在选择容器开发时需要。

应用本身通过 Lua 脚本实现验证码原子读取、删除，不依赖 GETDEL 命令。
依赖版本以 `go.mod`、`go.sum` 为准；升级 Go 不代表需要同时升级全部依赖。

## 快速开始

先获取仓库并进入项目根目录。以下命令使用 PowerShell，仓库地址由实际 GitHub 地址替换。

### 1. 准备依赖和配置

```powershell
go version
go mod download
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
```

`.env` 实际是 **YAML**，不是 `KEY=VALUE` 格式。编辑它，填写 MySQL、Redis 的实际地址和认证信息。已有 `.env` 时不要用示例覆盖自己的配置。

示例配置使用 `demo` 数据库。先通过数据库客户端创建数据库，并准备拥有所需权限的开发账号：

```sql
CREATE DATABASE demo CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
```

创建数据库、账号和授权属于数据库准备工作，迁移命令不会自动创建数据库。

HTTP 首次启动前生成持久化公私钥并创建加密版本备份：

```powershell
go run ./cmd/make make:auth-key
go run ./cmd/make make:auth-key-backup
```

重复执行保留已有私钥。公钥和加密私钥归档到 config/keys/<key_id>/，可提交 Git；明文私钥、备份口令留在被忽略的 storage/keys/。
请把备份口令另存到密码管理器，详见 [密钥恢复说明](docs/PASSWORD_ENCRYPTION.md)。新成员可生成自己的开发密钥，恢复原环境则使用对应版本备份和原口令。

### 2. 初始化表结构

```powershell
go run ./cmd/migrate -f ./.env
```

当前只有一份用户表初始化迁移：没有 `user` 表时创建完整结构，已有表时跳过。框架另外维护 `migrations` 执行记录表。
已有表被跳过不代表结构已被校验或补齐，详见 [用户表说明](docs/USER_SCHEMA.md)。

### 3. 验证代码

```powershell
go build ./...
go vet ./...
```

仓库应提交表映射、DO、DAO 底层和枚举生成代码，正常克隆不需要先连接数据库才能编译。
当前仓库按要求不保留测试文件，构建与静态检查通过不代表已执行业务测试。
表结构改变后才运行生成器，并提交对应生成文件：

```powershell
go run ./cmd/make make:dao -f ./.env -t user,migrations
gofmt -w model
```

### 4. 启动 HTTP

```powershell
go run ./cmd/api -f ./.env
```

默认示例监听 `:8088`，可访问：

- `GET http://127.0.0.1:8088/`：服务与模块信息。
- `GET http://127.0.0.1:8088/api/`：API 模块信息。
- `GET http://127.0.0.1:8088/system/ping`：后台模块存活检查。
- `GET http://127.0.0.1:8088/status`：数据库和 Redis 检查结果，需检查响应数据中的组件状态。

HTTP 启动入口叫 `cmd/api`，但同时提供 API 和 System 两个模块。队列消费者和定时任务需要在独立终端分别启动，见 [命令行说明](docs/COMMANDS.md)。

## 当前能力

| 能力 | 当前实现 |
| --- | --- |
| 路由与中间件 | Gin 的 httpx 封装、模块路由组、统一响应、追踪标识 |
| 用户认证 | RSA-OAEP密码字段密文、bcrypt、图片验证码、Redis Token、退出及账号状态检查 |
| 数据访问 | GORM/MySQL、生成的 entity/do/dao、模块独立 DAO |
| 迁移 | Go 文件注册、Up 执行、时间戳排序、执行记录 |
| 后台 | 独立目录与 `/system/ping`，尚未实现管理员登录和 RBAC |
| 异步任务 | Asynq、Redis、独立 worker、IP 白名单限制的队列监控 |
| 定时任务 | cron 独立进程和示例任务，尚无跨进程单实例调度锁 |
| 事件与外部服务 | event/listener 注册、rest 客户端示例 |
| 日志 | zerolog，HTTP/外部请求/队列/定时任务分类；尚未按 api/system 分文件 |

目前没有迁移回滚命令、请求/响应/控制器代码生成器、手机或邮箱验证、密码找回、资料编辑接口。不要把目录预留或示例当作已完成业务。

## 统一响应

正式入口通过 `config.InitGlobalVars()` 使用下列外层格式：

```json
{
  "status": 200,
  "msg": "成功",
  "data": {},
  "requestId": "请求追踪标识"
}
```

外层 `status` 是业务码。`data.status` 如果出现，是账号状态，两者含义不同。
业务错误通常仍为 HTTP 200，调用方必须同时判断外层 `status`；未登录使用 HTTP 401，限流使用 HTTP 429。
底层 httpx 未初始化时使用另一组默认字段，正式接口和接口测试应以入口初始化后的格式为准。

## 开发约定

新增功能按相同业务目录组织 `request → controller → logic → dao → response`，保持调用职责清晰，具体依赖关系见 [架构说明](docs/MODULES.md)。
数据库字段变更先写迁移，再执行、生成 DAO、更新请求与响应。已执行并共享的迁移不重写；用户端与管理端身份校验分别实现。

`.env.example` 可以提交；`.env`、环境私有配置、日志和二进制不提交。生成的 Go 代码随迁移一起提交，详见 [协作规范](docs/CONTRIBUTING.md)。
