# 模块与架构

## 应用形态

一个 Go module：`go-gin`。一个 HTTP 进程通过路径区分 `/api` 用户端和 `/system` 后台，共用配置、数据库连接、Redis 和基础组件。
这是单体应用内的模块化，不是多个独立部署的微服务；目录隔离和路径前缀不会自动产生权限隔离。

## 目录

```text
app/
├── api/
│   ├── controller/{auth,user}/  参数解析、业务调用和响应转换
│   ├── logic/{auth,user}/       认证与用户业务
│   ├── request/{auth,user}/     输入结构、绑定标签和校验规则
│   ├── response/{auth,user}/    对外输出结构
│   ├── dao/                    用户端专用数据访问
│   ├── transformer/            数据对象到响应对象的转换
│   ├── middleware/             API 身份校验、限流和模块标识
│   └── router/                 API 路由与中间件挂载
└── system/
    ├── controller/home/
    ├── logic/home/
    ├── response/home/
    ├── middleware/
    └── router/
cmd/{api,queue,cron,migrate,make}/ 独立程序入口
config/                          配置解析、全局响应格式、环境设置
router/                          全局路由、404/405、各端模块挂载
middleware/                      跨模块的全局中间件
model/
├── entity/                      数据库表映射，生成文件
├── do/                          构造查询/写入参数的数据对象，生成文件
└── dao/
    ├── internal/                DAO 底层，生成文件
    └── *.go                     可自定义扩展，生成器不覆盖已有文件
migration/                       当前完整 user 建表迁移
internal/                        httpx、数据库、Redis、队列、日志等组件
const/{errcode,enum}/             错误码与枚举
task/                            队列任务及注册
cron/                            定时任务及注册
event/                           事件、监听器及注册
rest/                            外部 HTTP 服务客户端示例
util/                            无业务归属的工具函数
```

System 当前没有输入结构，因此没有为了对称而建立空的 request 目录。后续出现请求时使用 `app/system/request/<feature>/`。
根目录遗留 `typing/query/.gitkeep` 仅为空占位，不是开发入口；不向旧 typing 目录添加代码。

## 启动过程

HTTP 的实际加载顺序：

1. `flag.Parse()` 解析 `-f`。
2. `config.Init()` 读取 YAML，使用 `sync.Once` 在进程内初始化一次。
3. `InitGlobalVars()` 设置 `status/msg/data/requestId`，`InitEnvironment()` 设置运行模式和时间配置。
4. 初始化日志 → MySQL → Redis → Asynq 投递客户端 → 事件监听器 → 外部服务客户端。
5. 创建 httpx 引擎，挂载恢复、追踪、访问日志、数据库检查和项目全局中间件。
6. `router.Init()` 显式挂载公共路由、`/api`、`/system`。
7. 入口挂载队列监控，监听 `app.port`。

业务路由不通过 `init()` 隐式挂载。迁移则沿用当前框架约定：迁移文件的 `init()` 注册自身，`cmd/migrate` 匿名导入 migration 包后统一执行。

## 请求链路

```text
HTTP 请求
  → 全局中间件
  → 模块中间件与路由认证
  → controller：解析 request
  → logic：业务规则、事务编排
  → dao：带 context 的数据库访问
  → response：显式选择输出字段
  → httpx：统一外层响应
```

`request` 和 `response` 是类型定义，不是每次请求必须执行的独立服务。没有数据库访问的业务无需额外造一个 DAO。
逻辑层接收 `context.Context` 和业务参数，不依赖 `*gin.Context`。身份来自中间件，客户端传入的用户ID不能替代当前身份。

## 模块边界

- `controller/user` 调用 `logic/user`；`request/user`、`response/user` 使用相同功能目录。
- `logic/user` 可复用 `logic/auth` 的登录态能力；业务包不反向导入 controller/router。
- `model/entity`、公共表 DAO 可共用；端专属查询放 `app/<端>/dao`，复杂后再按功能细分。
- 用户端查询需要约束当前用户或资源归属；后台查询需要管理员认证和权限校验。
- API Token 只服务 API。System 目前只有公开 ping，不能为新增敏感后台路由直接套用 API TokenCheck。
- 模块标记仅标识 `api/system`，当前不会自动按端分日志文件。

## 路由约定

顶层文件负责挂模块，模块内 router 负责挂功能。公开登录、注册和验证码与受保护路由分组。
用户路由在 `app/api/router/user.go` 内挂 `TokenCheck()`，保留 `/api/me`、`/api/logout`、`/api/user/`。
新功能的完整实现示例及 HTTP 方法约定见 [开发规范](DEVELOPMENT.md)。

`/api/demo/*` 仍是登录后可访问的示例，包含队列投递、事件和导出；不是正式业务契约。旧全表用户管理示例没有默认注册。

## 生成代码

迁移定义表，`make:dao` 读取表后生成 entity/do/dao 底层；它不创建表，也不反向生成迁移。
生成代码应提交仓库；修改应从迁移和生成器开始，定制查询放外层 DAO 或模块 DAO。
生成器保留已有外层 DAO，表被删除后不会自动清理旧文件，需要人工检查。
