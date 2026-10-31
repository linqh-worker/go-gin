# 命令行说明

以下命令均在项目根目录执行，要求 Go 1.26.1 或更高。配置的 MySQL 数据库与账号需要提前准备。
命令使用 Go 标准库 `flag` 和 `os.Args`，没有 Cobra，也没有统一 artisan 命令。

## 命令速查

| 用途 | 命令 | 外部依赖/写入 |
| --- | --- | --- |
| 下载依赖 | `go mod download` | 网络/Go 缓存，不更新依赖版本 |
| 持久化密码传输公私钥 | `go run ./cmd/make make:auth-key` | 本地文件，已有私钥不覆盖 |
| 加密版本备份 | `go run ./cmd/make make:auth-key-backup` | 公钥和加密私钥归档，口令独立保存 |
| 恢复公私钥 | `go run ./cmd/make make:auth-key-restore -i <备份文件>` | 不覆盖目标私钥 |
| 启动 HTTP | `go run ./cmd/api -f ./.env` | MySQL、Redis，监听 app.port |
| 启动队列消费者 | `go run ./cmd/queue -f ./.env` | MySQL、Redis，执行注册任务 |
| 启动定时任务 | `go run ./cmd/cron -f ./.env` | MySQL、Redis，执行注册任务 |
| 执行全部待执行迁移 | `go run ./cmd/migrate -f ./.env` | MySQL，创建/修改表并写迁移记录 |
| 创建迁移模板 | `go run ./cmd/make make:migration create_orders` | 仅创建 Go 文件，仍需填写 SQL |
| 生成指定表 DAO | `go run ./cmd/make make:dao -f ./.env -t user` | 读取 MySQL 表结构，写 Go 文件 |
| 生成多张表 DAO | `go run ./cmd/make make:dao -f ./.env -t user,migrations` | 同上 |
| 生成当前库全部表 DAO | `go run ./cmd/make make:dao -f ./.env` | 同上，包含该库其他表，需审查结果 |
| 生成枚举 | `go run ./cmd/make make:enum` | 读取 const/enum，生成 gen_*.go |
| 编译全部包 | `go build ./...` | 不启动服务，不连接数据库 |
| 静态检查 | `go vet ./...` | 不连接数据库 |

查看程序参数：`go run ./cmd/api -h`，queue、cron、migrate 同理。DAO 参数可通过 `go run ./cmd/make make:dao -h` 查看。
`go run ./cmd/make` 不带子命令会打印帮助并返回状态码1，不代表执行了生成。
`cmd/make` 有多个源文件，必须使用包目录执行，不要写成 `go run cmd/make/main.go`。

HTTP 不再依赖密码传输密钥。密钥命令仅供维护历史文件，备份恢复详见 [PASSWORD_ENCRYPTION.md](PASSWORD_ENCRYPTION.md)。

## 配置参数

支持 `-f` 的入口默认读取 `./.env`，也支持 `-f=./.env.testing`。

```powershell
go run ./cmd/api -f ./.env.testing
```

文件名可自定，但文件必须存在。内容为 YAML，运行模式取决于 `app.mode`，不会因文件名自动变为 testing。
没有通用 `--env`、`--conf`、`--port` 参数；端口在配置文件中设置。
生成器参数写在子命令之后：`make:dao -f ... -t ...`。`make:migration`、`make:enum` 不读取数据库配置。

## 三个常驻进程

需要哪些进程就分别开启对应终端，不必为了 HTTP 登录启动全部进程。

```powershell
# 终端一：HTTP，同时加载 api 和 system
go run ./cmd/api -f ./.env

# 终端二：异步任务消费者
go run ./cmd/queue -f ./.env

# 终端三：定时任务调度
go run ./cmd/cron -f ./.env
```

前台进程通过 `Ctrl+C` 停止；停止 HTTP 不会自动停止其他终端的进程。
HTTP 只初始化队列投递客户端，worker 才负责消费。Asynq 的任务在 Redis，不需要 MySQL jobs 表。
定时进程会加载 `cron/init.go` 的示例，多个副本会重复调度，当前没有跨进程调度锁。
不要仅靠删除 jobs DAO 或停止 HTTP 来判断队列是否停止。

## 迁移流程

### 新表

```powershell
go run ./cmd/make make:migration create_orders
```

打开新生成文件，设计表、字段、默认值、注释和索引，补完 `Up()`。模板的 TODO SQL 不能直接当作已完成迁移。
确认配置指向开发数据库后执行：

```powershell
go run ./cmd/migrate -f ./.env
go run ./cmd/make make:dao -f ./.env -t orders
gofmt -w model
go build ./...
```

### 已有表改字段

生成新的命名明确的迁移，例如 `make:migration add_orders_remark`，在 Up 中执行对应 ALTER。
已执行并共享的文件不改名、不重写；修改旧文件也不会使它自动重跑。
当前 user 初始化迁移仅 CREATE TABLE IF NOT EXISTS，不修复已有表的差异。

### 工具实际行为

- 按迁移结构体名称末尾14位时间戳排序，执行没有记录的迁移。
- 框架自动维护 `migrations` 表，用户不需要另外生成记录表的迁移。
- 尚无 rollback、status、fresh、reset、step 子命令，也没有自动反向导出迁移的命令。
- 当前入口部分错误仅打印后返回，进程退出码可能为0。执行后检查输出的失败信息及迁移记录，不能只看退出码。
- 当前没有并行迁移锁和跨 DDL 的事务保证；每个环境由一个迁移进程执行。
- 命令不会创建目标数据库，连接失败应先核对配置和服务状态。

## DAO 和枚举生成

`make:dao` 的方向是「数据库表 → Go 代码」。`-t` 使用实际表名，不是结构体名。

| 输出位置 | 重复生成行为 |
| --- | --- |
| model/entity | 覆盖，放表字段映射 |
| model/do | 覆盖，放查询/写入对象 |
| model/dao/internal | 覆盖，放底层方法 |
| model/dao/*.go | 已存在则跳过，保留自定义查询 |

生成器不会自动删除旧表对应文件。生成失败可能已经写入部分文件，应查看输出并复核后运行测试。
生成代码需要随业务提交到 Git，避免其他人克隆后缺包。
枚举生成也采用同样的提交策略：

```powershell
go run ./cmd/make make:enum
gofmt -w const/enum
```

目前没有生成 request、response、controller、logic 的命令，这些业务文件手写，并按 [开发规范](DEVELOPMENT.md) 放置。

## 构建检查

当前已清除项目测试文件和预置集成测试命令，使用以下命令检查编译与静态问题：

```powershell
go build ./...
go vet ./...
```

这些检查不启动服务、不访问数据库，也不能代替登录、注册、权限等业务验证。
原 GO_GIN_AUTH_INTEGRATION、GO_GIN_SCHEMA_INTEGRATION 对应的测试文件已删除，不再作为使用方式。

## 编译与运行

Windows PowerShell：

```powershell
New-Item -ItemType Directory -Path bin -Force
go build -o ./bin/api.exe ./cmd/api
go build -o ./bin/queue.exe ./cmd/queue
go build -o ./bin/cron.exe ./cmd/cron
go build -o ./bin/migrate.exe ./cmd/migrate
go build -o ./bin/make.exe ./cmd/make
./bin/api.exe -f ./.env
```

Linux/macOS，在相同目标平台编译：

```bash
mkdir -p bin
go build -o ./bin/api ./cmd/api
./bin/api -f ./.env
```

编译不会启动服务或执行迁移。部署二进制时一并准备配置，并固定工作目录。
`.devcontainer` 是可选环境。

## Air 开发热重载

在项目根目录执行：

```powershell
air -c .air.toml
```

配置会先构建 `./cmd/api` 到 `./tmp/api.exe`，再以 `-f ./.env` 启动。保存 Go 文件或默认 .env 时自动重新构建、启动；日志、文档和 tmp 不触发构建。
Air 自己的 `-c` 指定 Air 配置，应用的 `-f` 则放在 `.air.toml` 的 `build.args_bin` 中，两者不是同一个参数。

改用其他应用配置文件时，将 .air.toml 中 args_bin 改为 `["-f", "./.env.testing"]`，并将该文件加入 include_file，修改后重启 Air。
首次安装或按当前已验证版本安装：

```powershell
go install github.com/air-verse/air@v1.65.1
```

安装后确保 Go 的 bin 目录位于 PATH，可用 `air -v` 查看。本机当前已安装，无需重复安装。
Air 仅用于开发，退出按 Ctrl+C；不要同时运行普通 go run 和 Air 占用同一端口。

创建管理员：设置 SYSTEM_ADMIN_PASSWORD 环境变量后执行 go run ./cmd/admin -f ./.env -username admin；完整步骤见 [后台认证](SYSTEM_AUTH.md)。

## 合并运行与按端运行

同一份程序通过启动参数选择模块；不传 module 时默认 all，保留原来的单端口方式。port 未传时使用 .env 的 app.port；传入时覆盖本次进程监听地址，不修改配置文件。

```powershell
# 本地合并启动
go run ./cmd/api -f ./.env --module=all --port=:8088
# 分别在两个终端启动
go run ./cmd/api -f ./.env --module=api --port=:8088
go run ./cmd/api -f ./.env --module=system --port=:8089
```

Linux 服务器使用与目标系统匹配的可执行文件：

```bash
# Linux 环境中编译
go build -o go-gin ./cmd/api
# 合并运行（选择一种部署方式即可）
./go-gin -f ./.env --module=all --port=:8088
# 或分别以两个进程运行
./go-gin -f ./.env --module=api --port=:8088
./go-gin -f ./.env --module=system --port=:8089
```

Windows 下交叉编译 Linux amd64：

```powershell
$env:GOOS = 'linux'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'
try { go build -o ./go-gin ./cmd/api } finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
}
```

目标机器为 ARM64 时将 GOARCH 改为 arm64。启动时配置文件相对于工作目录解析，服务器应配置正确的工作目录。

两端保留 /api、/system 路径前缀，分端后不会变成 /login。/ 和 /status 是每个进程共有的公共接口。队列监控只在 all/system 模式挂载，继续使用原 IP 白名单。

两个进程可共用数据库、Redis和配置文件，各自持有独立连接池；验证码、Token 和限流仍按端隔离。分端监听地址必须不同；HTTP 参数不会启动队列消费者或定时任务进程，它们继续按原命令独立运行。

授予已有管理员超级管理员角色：go run ./cmd/admin -f ./.env -username admin --grant-super。需要先执行 RBAC 迁移，详见 [RBAC 文档](SYSTEM_RBAC.md)。

补齐已有业务表注释：go run ./cmd/migrate -f ./.env。CompleteTableComments20261031220000 已覆盖用户、管理员、RBAC及菜单关联表；数据库客户端需刷新表结构才能显示最新注释。
