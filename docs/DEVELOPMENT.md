# 项目开发流程说明

本篇回答“一个功能从哪里开始、按什么顺序写”。固定规则见[开发约定](AGENTS.md)，启动、配置和接口调用见[使用文档](USAGE.md)，整体架构见 [README](../README.md)。

## 1. 先确定功能边界

开始前明确：属于用户端api还是管理端system，输入输出是什么，是否登录、需要访问哪张表、能操作谁的数据、成功和失败是什么。

真实请求顺序是：

```text
router → 本端middleware → controller → logic → DAO → MySQL
                                ↑          ↓
                       响应映射 ← 结果或错误
```

request和response是输入输出的数据定义，不是额外运行的服务层。logic承担service职责；auth_error等文件存放错误及转换，不是每次请求依次进入的下一层。

开发可以先定义request/response，再写logic、DAO、controller，最后注册router。简单功能不用复制所有层：不查库就不用DAO，没有业务复用也不用新接口包装。

## 2. 有表结构变化时先做迁移

以下以未来的orders表为例，命令在项目根目录执行：

```powershell
go run ./cmd/make make:migration create_orders
```

生成器仅创建模板。打开 `migration/` 新文件，填写Up里的建表SQL，设计字段类型、默认值、NULL、主键、唯一索引、关联和中文注释，再确认配置指向独立开发/测试库：

```powershell
go run ./cmd/migrate -f ./.env.testing
go run ./cmd/make make:dao -f ./.env.testing -t orders
gofmt -w model
```

执行顺序是“迁移定义 → 数据库表 → 生成Go映射”。生成DAO不会创建表，也不会从数据库自动导出迁移。已有表新增字段时另建 `add_orders_remark` 等增量迁移；修改已执行文件不会自动重跑。

| 生成位置 | 用途 | 重复执行 |
| --- | --- | --- |
| model/entity | 表结构映射 | 覆盖 |
| model/do | 查询条件/写入对象，字段为any | 覆盖 |
| model/dao/internal | 表名、列名、Ctx、事务封装 | 覆盖 |
| model/dao/*.go | 外层DAO，可补定制查询 | 已存在则跳过 |

生成后查看diff，提交生成代码。需要修改生成结构时改表和生成来源，不直接手改内部生成文件；被删除表的旧生成文件需另行审查清理。

## 3. 定义request和response

使用业务目录，例如用户资料放 `request/user`、`response/user`。输入只包含用户允许提交的字段，输出只包含允许公开的字段。

更新接口用指针区分“未提交”和“提交零值”：

```go
type UpdateReq struct {
    Nickname *string `json:"nickname" binding:"omitempty,max=50"`
    Age      *int    `json:"age" binding:"omitempty,gte=0,lte=150"`
}
```

`Age == nil` 表示不修改；`Age != nil && *Age == 0` 表示明确更新为0。账号状态、角色分配应使用专门接口，不混入普通资料请求。

## 4. DAO沿用生成器工作流

在模块DAO内组合生成的公共DAO和DO，例如：

```go
package dao

import (
    "context"
    modeldao "go-gin/model/dao"
    "go-gin/model/do"
    "go-gin/model/entity"
)

func FindEnabledUser(ctx context.Context, id int) (*entity.User, error) {
    var row entity.User
    found, err := modeldao.User.Ctx(ctx).
        WherePri(id).
        Where(do.User{Status: 1}).
        WhereNull("deleted_at").
        Found(&row)
    if err != nil { return nil, err }
    if !found { return nil, nil }
    return &row, nil
}
```

不能将返回entity直接输出到HTTP；由logic/controller转换为response。用户私有资源还要加当前用户ID或资源归属条件。

零值更新示例：

```go
rows, err := modeldao.User.Ctx(ctx).
    WherePri(id).
    Data(do.User{Age: 0}).
    Update()
```

DO字段名以生成结果为准，如Id、AvatarUrl、RegisterIp；nil表示不参与写入。新增需要ID时用 `InsertAndGetId()`，Insert返回的是影响行数。

事务调用示例，放在实际需要一致性的业务/DAO中：

```go
return db.Transaction(ctx, func(tx *db.TX) error {
    var current entity.User
    found, err := modeldao.User.Ctx(ctx).WithTX(tx).
        WherePri(id).LockUpdate().Found(&current)
    if err != nil { return err }
    if !found { return errorx.New(10001, "用户不存在") }
    _, err = modeldao.User.Ctx(ctx).WithTX(tx).
        WherePri(id).Data(do.User{Nickname: nickname}).Update()
    return err
})
```

这里涉及的db、entity、do、errorx和modeldao分别从框架及model包导入。所有事务语句使用同一个tx，具体业务还需检查状态、归属等条件。

## 5. 写logic，再由controller调用

logic处理业务规则、调用DAO或rest，返回数据/错误；controller只做请求处理和响应边界。logic不导入Gin，也不直接写HTTP JSON。

下面给出可按文件手写的简单例子：新增登录后的问候接口。它不查库，因此无需迁移或DAO。

### 请求：app/api/request/user/greeting.go

```go
package user

type GreetingReq struct {
    Prefix string `form:"prefix" binding:"max=20" label:"问候前缀"`
}
```

### 响应：app/api/response/user/greeting.go

```go
package user

type GreetingResp struct {
    Message string `json:"message"`
}
```

### 业务：app/api/logic/user/greeting_logic.go

```go
package user

import (
    "context"
    "strings"
    userRequest "go-gin/app/api/request/user"
    userResponse "go-gin/app/api/response/user"
)

type GreetingLogic struct{}

func NewGreetingLogic() *GreetingLogic { return &GreetingLogic{} }

func (l *GreetingLogic) Handle(ctx context.Context, nickname string, req userRequest.GreetingReq) (*userResponse.GreetingResp, error) {
    prefix := strings.TrimSpace(req.Prefix)
    if prefix == "" { prefix = "你好" }
    return &userResponse.GreetingResp{Message: prefix + "，" + nickname}, nil
}
```

### 控制器：app/api/controller/user/greeting_controller.go

```go
package user

import (
    userLogic "go-gin/app/api/logic/user"
    "go-gin/app/api/middleware"
    userRequest "go-gin/app/api/request/user"
    userResponse "go-gin/app/api/response/user"
    "go-gin/internal/errorx"
    "go-gin/internal/httpx"
    "go-gin/internal/httpx/validation"
)

type greetingController struct{}
var GreetingController = &greetingController{}

func (c *greetingController) Show(ctx *httpx.Context) (any, error) {
    var req userRequest.GreetingReq
    if err := ctx.ShouldBindQuery(&req); err != nil {
        return nil, validation.Error(err, &req)
    }
    raw, exists := ctx.Get(middleware.CurrentUserKey)
    current, ok := raw.(*userResponse.AuthUser)
    if !exists || !ok || current == nil {
        return nil, errorx.NewServerError(401)
    }
    return userLogic.NewGreetingLogic().Handle(ctx.Request.Context(), current.Nickname, req)
}
```

### 路由：app/api/router/user.go

在现有private组添加；该文件已有controller作为user控制器包的别名：

```go
private.GET("/user/greeting", controller.GreetingController.Show)
```

完整路径为 `GET /api/user/greeting?prefix=早上好`，请求携带API Bearer Token。示例尚未加入业务源文件，不能在未实现时直接调用。

## 6. 注册新功能和后台权限

同一功能继续放现有业务包；新功能如orders新增模块router文件并从该端router/init.go调用注册函数。

API公开接口放公开组，私有接口复用该端TokenCheck。System模块入口已统一挂Authorize，新敏感接口不再逐个套Token中间件。

后台新增接口还需要：

1. 注册明确的HTTP方法与路径，如 `GET /system/orders/:id`。
2. 在对应页面下创建button或api菜单节点，填写method、api_path完整模板。
3. 给角色关联该节点，再给管理员关联角色。
4. 分别验证获授权、未授权、不同方法、动态ID、角色/父菜单停用和撤权。

只加路由会默认403，超管也必须有有效的菜单配置。新增“永久基础功能”的节点应通过新数据迁移提供，不能只在本机数据库点选后就认为别人克隆也有配置。公开/个人/超管边界维护在 `app/system/authorization/routes.go`，只有业务明确需要时才修改。

## 7. 接入任务、事件和其他系统

- 需要异步执行：task中定义任务和处理器，在task/init.go注册，HTTP进程投递、queue进程消费；参考现有SampleTask。
- 需要周期执行：在cron定义job并由cron/init.go注册，独立启动cmd/cron，先调整现有示例调度。
- 需要事件：event定义事件并由event/init.go关联listener。
- 调用第三方/其他业务系统：在rest按服务封装HTTP客户端，由logic调用，配置在svc；它与本项目接收请求的controller方向不同。

先明确幂等、重试、错误和日志；需要哪些进程再启动哪些，登录注册不依赖运行cron或worker。

## 8. 验证与交付顺序

```powershell
# gofmt只处理本次修改的Go文件，以下按实际路径替换
# gofmt -w app/api/controller/user/greeting_controller.go

go build ./...
go vet ./...
git diff --check
git status --short
```

按实际变更验证成功/失败、缺失/无效身份、越权、重复请求、零值、停用及事务失败。数据库验证使用隔离环境，不能依赖被测ID直接清理真实数据。

当前不保留永久测试文件，也没有可直接复跑139项HTTP验证的命令。临时验证按任务范围编写、执行后清理并记录结果；过去通过不代替新变更验证。

交付时同时提供源代码、迁移、生成代码和文档，写明需要先执行的迁移及行为变化；命令与接口更新到USAGE，稳定规则更新到AGENTS，能力变化更新README。部署先迁移再启动新版本，详见[使用文档](USAGE.md)。
