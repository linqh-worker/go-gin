# 统一业务开发规范

本规范用于后续新增功能与被修改的旧代码。保持已有对外接口兼容；目录统一不等于每次都要重构全部存量代码。
只有实际需要时才新增层、接口和目录，不建立没有业务用途的空包。

## 1. 目录与命名

```text
app/<module>/
├── controller/<feature>/   HTTP 请求入口
├── logic/<feature>/        业务规则与流程
├── request/<feature>/      请求参数、绑定标签和格式校验
├── response/<feature>/     对外响应字段
├── dao/                    端专属的数据访问，需要时再细分 feature
├── transformer/            数据到响应的转换，需要时使用
├── middleware/             本端身份、权限、限流等
└── router/                 路由与中间件注册
```

- module 使用 api、system 等端名称；feature 使用 user、auth、order 等业务含义。
- 手写文件使用小写 snake_case，如 update_profile_logic.go；同目录 package 名使用对应功能，如 user。
- 请求类型以 Req 结尾，响应以 Resp 结尾；现有 AuthUser 等嵌套数据类型保留，不强行改对外字段。
- 新手写字段使用 ID、URL、IP 等 Go 常见缩写；生成代码遵守生成器输出，不手动批量改名。
- 功能可有多个文件，不把注册、资料、订单等所有业务塞进一个大文件。
- 同名功能包导入时明确别名，如 userRequest、userResponse、userLogic、userController。
- 不使用旧 typing 目录，不在项目根目录新建散落的 controller、logic 或重复的公共框架。

## 2. 各层职责

| 层 | 应做 | 不应做 |
| --- | --- | --- |
| router | HTTP 方法、路径、认证与权限中间件挂载 | 编写 SQL 或业务流程 |
| controller | 解析 request、读取可信身份、调用 logic、映射返回错误 | 实现事务、密码规则、余额或状态流转 |
| request | 明确输入字段，json/form/uri 标签，格式和范围校验 | 当作数据库实体，接收客户端权限和可信身份 |
| logic | 校验业务条件、权限范围、状态流转、事务与外部调用编排 | 依赖 Gin Context 或在业务中注册路由 |
| dao | 带 context 的查询、写入与数据库约束错误转换 | 直接返回 HTTP JSON 或信任请求中的当前用户ID |
| response | 定义允许公开的字段和嵌套结构 | 导出密码哈希、完整内部状态或数据库实体全集 |
| transformer | 明确的数据到响应映射 | 隐式查库或改变业务状态 |

数据库表对应 entity，查询/写入对象对应 do；它们与 request/response 不要求字段相同。
请求体经校验后显式映射到写入字段，不能直接把整个 JSON map 交给数据库 Updates。

## 3. 请求与响应

- JSON 写入使用显式 json 标签；GET 查询使用 form 标签；路径参数使用 uri 标签并单独绑定。
- 必填、长度、范围、枚举等格式在 request 校验；用户名唯一、资源归属、账号状态、库存等在 logic/数据库保证。
- 当前用户名登录请求至少8字符密码，bcrypt 输入不超过72字节；不能因加字段而绕过已有校验。
- 登录注册直接接收密码字符串，不要求 encryption；数据库保留 bcrypt 哈希，部署使用 HTTPS。
- 当前用户ID来自 TokenCheck 写入的 user_id 或当前用户对象；客户端字段不能充当身份来源。
- 可信IP、账号状态、权限、创建者、密码哈希由服务器赋值。不要新增接收这些字段的普通用户参数。
- 分页接口统一采用 page（默认1）、page_size（默认20，最大100）；新接口在 request/logic 中实现默认值及上限，不假设框架自动分页。
- 列表响应按需提供 items、total、page、page_size；空列表约定返回 []，并在转换处初始化切片。现有旧接口的 data 字段不擅自改名。
- JSON 新字段统一 snake_case；外层 requestId 是既有兼容字段，保留当前形式。

正式入口的外层格式固定为 `code/msg/data/requestId`。成功业务码为200，业务错误通常 HTTP 200，需要同时判断外层 code。
401 表示身份失效，403 表示权限不足，404 表示资源或路由不存在，429 表示限流；基础设施异常使用500/503。
data.status 如果是账号状态，需要在接口文档中与外层 code 区分。
控制器返回 `(any, error)`，交给 httpx 统一写响应；不要既 ctx.JSON 又返回普通结果导致写两次。文件下载等流式响应沿用对应组件的专用处理方式。

## 4. 错误处理

可预期业务失败使用已定义的业务错误。新增稳定错误码放 const/errcode，先检索已有编号，避免与当前20001至20007等冲突。
内部操作保留原始 cause 便于 errors.Is/errors.As；对客户端转换为已知消息，不直接暴露 SQL、Redis 错误或第三方响应。
现有认证通过 auth.HTTPError 做边界映射；新业务使用自己的明确错误映射，不拿认证错误覆盖所有业务。
不要忽略数据库写入、事务提交、Redis 登录态创建等关键错误。补偿失败需记录安全的诊断信息，不能假装操作完整成功。

## 5. 身份与数据权限

- 公开接口和私有接口使用不同路由组，私有组先挂身份中间件再注册 handler。
- 用户端身份不能作为管理员身份，System 新增敏感功能须先实现自己的管理员认证和权限。
- 数据查询/更新同时约束资源ID和归属用户，不能只在页面隐藏按钮。
- 账号停用和软删除继续在认证时检查，不能只相信 Token 曾经签发过。
- 金额、价格、角色、资源拥有者等由服务端计算或校验，不相信客户端提交的结果。
- 退出、修改、删除等有状态操作不用 GET，不随意使用 Any 注册正式业务。
- `/api/demo/*` 和旧用户管理 logic 是示例，新增正式业务不要直接复制其 GET 写入、无分页全表查询等模式。

## 6. 数据库与迁移

新表流程为「生成迁移模板 → 设计 Up SQL → 执行开发库迁移 → 生成指定表 DAO → 开发业务 → 测试」。
当前 user 基线只负责创建一张完整用户业务表，已有表跳过；框架维护 migrations 记录。后续改字段新增增量迁移。

- 表名、字段、索引及注释明确；项目不要求应用前缀，后台表可用 system_admins 等模块语义命名。
- 当前 user 是历史单数表名，不为了新命名规范擅自改名；新表可以采用业务含义明确的小写 snake_case。
- 必填和可空要区分；手机号/邮箱的空值和唯一性按实际验证流程设计，不以空串默认值代替所有 NULL。
- 存金额用整数最小货币单位或明确精度的 DECIMAL，不用 float64 进行资金计算。
- 唯一性由数据库约束保证，提前查重只能改善提示，不能替代唯一索引。
- 查询传递 context，SQL 使用参数化；动态排序、表名、列名不能直接拼接用户输入，需白名单。
- 使用稳定排序和分页，避免正式列表接口无上限读取全表。
- 多表一致性在 logic 编排事务；事务中的 DAO 必须使用同一 tx，不能中途回到全局连接。
- 并发计数使用数据库原子自增或适当锁，状态变化用带原状态的条件更新并检查影响行数。
- 外部 HTTP、支付和队列投递不天然与数据库事务原子一致，需要按业务设计幂等键、补偿或 outbox，不能把网络调用塞进长事务当成一致性保证。
- 当前迁移工具没有回滚/并发锁，不并行执行多个迁移进程，执行后检查日志与记录。

生成的 entity/do/dao/internal 和枚举文件随迁移提交；不直接编辑它们实现业务。定制代码放外层 DAO 或模块 DAO。
生成器不会删除旧文件，也可能只完成部分输出；生成后查看变更并运行 gofmt、测试。

## 7. 日志、任务与外部服务

业务日志使用 logx.WithContext(ctx)，记录操作、资源ID、结果和必要的错误，不记录密码、Token、验证码答案、完整请求头或配置。
用户输入可以包含隐私和控制字符，不把整份请求体直接打印。

队列任务定义在 task，注册在 task/init.go；事件监听在 event/init.go 注册；定时任务在 cron/init.go 注册。
异步任务应带必要的业务ID，设计重复执行不会重复扣款、发货或插入的幂等行为；不能仅凭任务已投递就向用户宣称业务已完成。
当前 Asynq 使用 Redis，不依赖 MySQL jobs 表。当前 cron 没有跨副本锁，同一环境保持单调度进程。
rest 客户端按服务封装，明确超时、重试条件和错误映射；对非幂等操作不要无条件重试。

## 8. 示例：新增当前用户问候接口

以下是开发步骤示例，文件并未自动加入现有业务。这个接口只读取中间件已验证的用户，不查数据库，因此无需为它创建迁移或 DAO。

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
    "go-gin/const/errcode"
    "go-gin/internal/errorx"
    "go-gin/internal/httpx"
)

type greetingController struct{}
var GreetingController = &greetingController{}

func (c *greetingController) Show(ctx *httpx.Context) (any, error) {
    var req userRequest.GreetingReq
    if err := ctx.ShouldBindQuery(&req); err != nil {
        return nil, errcode.NewDefault("问候前缀最多20个字符")
    }
    value, exists := ctx.Get(middleware.CurrentUserKey)
    current, ok := value.(*userResponse.AuthUser)
    if !exists || !ok || current == nil { return nil, errorx.ErrUnauthorized }
    return userLogic.NewGreetingLogic().Handle(ctx.Request.Context(), current.Nickname, req)
}
```

### 路由

在 app/api/router/user.go 的现有 private 组中添加，controller 已是 user 控制器包的别名：

```go
private.GET("/user/greeting", controller.GreetingController.Show)
```

完整路径为 GET /api/user/greeting，继承 TokenCheck，不把身份校验重复写进 logic。
对更复杂的资源变更，再补充模块 DAO、事务、数据归属与相关测试，不要为这个简单例子额外复制框架。

## 9. 测试与完成标准

- 新业务测试优先与对应包同目录，测试名称说明行为，而不是只验证自己写出的字段排列。
- 认证、权限、唯一性、状态变化、并发写入等修改必须验证拒绝路径和失败后的状态。
- 当前测试文件已按要求清除，不自动恢复；后续安排自动化测试时，优先使用内存依赖或模拟仓储，真实库验证只使用开发库并清理临时数据。
- 请求或响应字段变化同步接口文档；表结构变化同步迁移、生成代码和字段说明。
- 使用 gofmt，完成 go build ./... 与 go vet ./...；报告实际业务验证，不把编译成功当作测试覆盖。
- 交付说明写清改了什么、验证了什么、是否需要执行迁移、是否影响调用方。未完成的功能明确列出，不以目录或模板代替实现。

完整命令见 [COMMANDS.md](COMMANDS.md)，提交规则见 [CONTRIBUTING.md](CONTRIBUTING.md)。

公共绑定错误使用 `internal/httpx/validation.Error(err, &req)`；需要验证码等专属提示时，先用 `validation.Normalize(err, &req)` 获取字段错误，匹配专属规则后，其余交给 `validation.Error(err, nil)`。此转换属于 controller 请求处理，不放 logic。

新建业务表必须为表和每个字段添加中文 COMMENT，说明状态枚举、关联ID、空值和默认值含义。现有表注释补全由 CompleteTableComments20261031220000 执行，读取原列定义后仅更新注释。
