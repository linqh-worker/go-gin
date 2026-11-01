# 项目开发约定

人工与AI修改本项目共同遵守本文件。架构见 [README](../README.md)，实现步骤见[开发流程说明](DEVELOPMENT.md)，命令及接口见[使用文档](USAGE.md)。本文件位于docs，工具不会作为根目录AGENTS自动加载，开始修改前应明确阅读。

## 1. 目录与职责

- Go module为 `go-gin`，最低Go版本1.26.1，未经需求不整体升级依赖或修改module。
- 模块使用api、system，URL前缀分别为 `/api`、`/system`；分端口不改变路径。
- 按 `app/<module>/{controller,logic,request,response}/<feature>` 分功能。端专属查询放 `app/<module>/dao`，有实际需要再细分。
- 文件小写snake_case，请求以Req结尾，响应以Resp结尾；手写字段使用ID、URL、IP缩写。生成代码遵守生成器命名。
- 同名包导入明确别名，如accountRequest、accountResponse、accountLogic。
- logic就是service职责，不再套同职责层或无实际用途的接口。公共工具按实际需要提取，不建立空目录。
- 不恢复已删除的demo路由、examples、旧typing和根公共middleware；业务中间件保留在各端内部。
- 仅维护README、DEVELOPMENT、AGENTS、USAGE四份主文档，变化更新对应章节。

| 层 | 应承担的职责 |
| --- | --- |
| router | 方法、路径、分组及中间件挂载 |
| middleware | 本端身份、权限、限流及上下文 |
| controller | 绑定参数、获取可信身份、调用业务、转换响应及错误 |
| request | 输入字段、格式、长度及范围校验 |
| logic | 业务条件、状态流转、事务及外部调用编排 |
| dao | 数据访问、写入映射、数据库约束错误转换 |
| response | 明确允许公开的输出字段 |
| transformer | 必要时集中做映射，不隐式查库 |

logic/DAO接收 `context.Context`，不依赖Gin、controller或router。controller传 `ctx.Request.Context()`。错误文件只是定义及转换，不是每次请求必经的额外执行层。

## 2. 数据访问与迁移

- 业务使用生成的 `model/dao` 配合 `model/do`，app和管理员命令不直接拼表名写SQL。迁移DDL及基础数据脚本可使用明确的参数化SQL。
- 定制查询放外层 `model/dao/*.go` 或模块DAO。不直接修改entity、do、dao/internal、枚举生成文件实现业务。
- 新表/字段先设计迁移，执行指定开发库后生成对应表代码。已执行并共享的迁移不改名、不删除、不重写；后续新增增量迁移。
- 初始化只 `CREATE TABLE IF NOT EXISTS`，不能自动删除、清空或重建已有数据。已有表差异用增量迁移修复。
- 新表及每个字段添加中文COMMENT，明确NULL、默认值、状态、关联ID。唯一性由数据库唯一约束兜底。
- 当前沿用user、system_admin等单数表名，不批量改名；不要求统一应用前缀，表名前缀不提供数据权限。
- 金额用最小货币单位整数或明确精度DECIMAL，不使用float表示资金。
- DO的nil字段不参与条件/写入，零值必须显式赋值；DO字段使用生成器命名，如Id、AvatarUrl，避免缩写转换成错误列名。
- `Insert()` 返回影响行数；需要主键使用 `InsertAndGetId()`，不能将影响行数当主键。
- 多表操作使用同一事务，内部DAO必须 `.WithTX(tx)`；需要行锁用 `.LockUpdate()`，事务中不能回到全局连接写入。
- 并发计数使用数据库原子更新，如gorm.Expr；状态更新约束必要原状态并检查影响行数。
- 查询有稳定排序及分页；排序字段等动态条件经过白名单，用户输入不直接拼SQL。
- 用户资源查询约束当前身份和资源归属；数据库实体不能直接作为正式HTTP响应。
- 外部HTTP、Redis和数据库不是同一事务，按业务设计幂等和补偿，不用长事务包网络请求。
- 迁移同一环境只运行一个进程，检查输出与migrations记录；部分失败退出码不可靠，不能只看退出码。

## 3. 请求、响应和错误

- JSON字段显式定义；GET使用form标签；路径ID单独校验为正整数。
- 格式校验在request/controller；唯一性、账号状态、库存、资源归属等业务规则在logic/数据库。
- 写入显式映射白名单，不把整份JSON map传给Updates。客户端不能提交可信身份、IP、密码哈希或角色，也不能借资料接口修改状态。
- 用户名3至32位字母数字下划线，转小写；密码至少8字符、UTF-8最多72字节；数据库存bcrypt。
- 外层固定 `code/msg/data/requestId`，成功code=200。data.status是账号状态，两者不同。
- 列表按需使用items/total/page/page_size；page默认1，page_size默认20且最多100；空数组返回[]。
- 新JSON字段用snake_case，requestId保持兼容。用户时间用 `util.FormatTime`，北京时间 `2006-01-02 15:04:05`，空时间为null。
- controller返回 `(any,error)` 交给httpx，不同时手写JSON又返回普通结果。文件流使用组件专用方式。
- 参数错误用 `validation.Error(err,&req)`；专属规则先Normalize，再交给公共转换。
- 业务失败使用可识别错误，在模块边界转换，保留cause支持errors.Is/As；客户端不暴露SQL、连接信息或内部调用响应。
- 稳定错误码新增前检查const/errcode及业务包现有编号。不忽略写库、事务提交、签发Token或补偿失败。

## 4. 认证与后台权限

- API使用本端TokenCheck；System在模块入口统一Authorize，不在控制器重复解析Bearer或逐路由复制鉴权。
- 用户与后台账号、验证码、Token和限流独立，API身份不能作为管理员身份。
- System只公开GET captcha及POST login；GET me、GET menus/mine和POST logout要求登录；其他业务接口默认拒绝。
- 菜单与接口权限统一保存在system_menu，使用method/api_path；角色只关联system_role_menu。不重新添加权限编码、独立接口表或接口规则表。
- 按请求方法和 `c.FullPath()` 匹配模板，如 `/system/users/:id`，不能直接拿实际ID路径对比模板。
- 节点与全部祖先必须启用；普通管理员须直接关联对应节点，超管也不能绕过缺失/停用配置。
- 授权管理、菜单维护及管理员写操作保留超管限制。HTTP不能授予超管或修改保留超管角色关系。
- 显示补齐祖先只用于展示；勾父节点不自动获得全部后代，前端必须提交实际勾选节点ID。
- 每次认证检查账号、角色与菜单；撤权即时生效。停用再启用不表示旧Token已物理撤销。
- 菜单接口权限不自动提供部门、租户及资源归属过滤，数据范围需在业务查询中实现。

## 5. 配置、日志与任务

- `.env` 是YAML，示例只放占位值。通过-f指定完整文件，改配置重启相关进程；不假定系统环境变量能覆盖所有配置。
- 当前不要求请求体密码加密，部署使用HTTPS。历史密钥工具不参与HTTP认证。
- 日志用 `logx.WithContext(ctx)` 关联追踪及api/system模块；controller向DAO传Request.Context，禁止请求途中切换全局日志模块。日志按端及app/access/sql/httpc分类，SQL使用WithSQLContext并保留参数占位符。不记录密码、完整Token、验证码答案、密钥、整份请求头或配置。
- rest按外部服务封装客户端，明确超时、错误映射和重试条件，非幂等请求不无条件重试。
- 任务在task定义并由task/init.go注册；事件在event/init.go注册；定时任务在cron/init.go注册。
- 异步任务设计重复执行安全；投递成功不代表业务执行完成。当前cron保持单调度实例。

## 6. 验证和数据保护

- 修改前检查现有代码和工作区，保留用户改动，不自动恢复被删除的测试或示例。
- 纯文档修改检查链接与命令，不运行数据库操作。业务修改完成gofmt、go build ./...、go vet ./...。
- 当前按要求不保留测试文件；如需临时行为验证，报告实际范围，不能把编译通过当成功能全部通过。
- 鉴权、唯一性、并发、资金和状态变化验证成功、拒绝及失败后状态。真实写入优先使用独立临时MySQL库和隔离Redis。
- 清理前核对环境、唯一业务标识及实际记录，不能只依赖被测返回ID删除数据。临时库使用生成名称、固定前缀并校验后清理。
- 有条件时核对原库验证前后指纹；临时程序和数据结束后清理。备份进入忽略目录，不提交真实数据。

## 7. 交付与版本控制

提交业务源码、迁移、生成文件、go.mod/go.sum、公开配置模板和四份主文档。不提交私有环境文件、日志、二进制、明文密钥、备份口令、真实账号与数据库备份。

`config/keys/<key_id>/` 只保存公钥和加密私钥归档；口令独立备份。生成文件正常提交，不用git add -f。

提交前检查git status、git diff --check及暂存内容；说明描述具体改动，如 `feat(user): 增加用户资料维护`。交付写清行为、验证、迁移及部署顺序和未解决事项。未经用户要求不自动提交或推送Git。

前端可使用既有system_menu.code作为按钮显隐标识，不增加permission_code字段或独立权限表。按钮使用公共hasPermission和v-if；权限只来自直接关联的启用节点，展示补齐的祖先不产生额外授权，后端继续按请求方法和接口模板鉴权。变更现有菜单code需同步前端引用，并用增量迁移兼容已有库。

演示管理员保护以system_admin.is_protected为准，HTTP不得修改该标记；资料、密码、状态和角色写入均须在锁定账号后校验。管理员改密仅超管可用，校验确认密码、存bcrypt并使已有后台Token失效。

系统动态配置保存在system_config；连接凭据及私钥不进入配置管理。公开读取只允许is_public=1，后台写入只接受key/value/version，按组事务校验并记录操作者，不允许通过保存值接口修改配置元信息或公开范围。配置定义由独立的超管接口维护，不能通过值保存接口修改。内置项不可删除、改key、移动分组、改数据类型或取消保护。分组采用同表record_type=group记录，公开查询必须同时约束record_type=item与is_public=1。配置写入先按ID锁分组行再锁配置项，结构变动同步版本。添加默认配置用增量迁移且不覆盖已设置值。
