# 用户表说明与表名前缀

## 当前命名约定

Go 不要求表名前缀。表名是项目和数据库的约定，不是语言限制。
本项目沿用现有 `user` 表，API 中仍使用 `username`，数据库列仍为 `name`。
以后后台管理员建议使用独立的 `system_admins` 表；业务表按订单、商品等实际含义命名。

如果多个独立应用必须共享一个数据库，可以约定 `shop_users`、`shop_orders` 等应用前缀。
如果项目有独立数据库，通常不需要为每张表重复添加应用前缀。
同一个业务模块的表可以采用统一语义前缀，例如 `system_admins`、`system_roles`、`system_permissions`。
前缀不提供身份认证、数据权限或数据库访问隔离。

本项目 DAO、生成器和部分 SQL 使用显式表名；以后若要改名，需通过迁移改实际表名并同步 DAO、SQL 和生成代码。
不能只设置 GORM 的 TablePrefix 就认为所有原生 SQL 和已生成的 TableName 都会跟着改变。

## user 字段

| 字段 | 类型 | 用途 |
| --- | --- | --- |
| id | INT 自增 | 主键 |
| name | VARCHAR(100) | 登录用户名；接口约束为3至32位字母、数字、下划线，转小写 |
| password_hash | VARCHAR(100)，可空 | bcrypt 密码哈希；旧账号没有密码时为空 |
| nickname | VARCHAR(50)，默认空串 | 昵称，新注册默认使用用户名 |
| avatar_url | VARCHAR(255)，默认空串 | 头像地址；不包含上传功能 |
| phone | VARCHAR(32)，可空 | 手机号资料，保存不代表已验证 |
| email | VARCHAR(254)，可空 | 邮箱资料，保存不代表已验证 |
| gender | TINYINT UNSIGNED，默认0 | 0未知、1男、2女 |
| age | INT，可空 | 保留已有年龄字段，兼容原示例 |
| status | INT，可空，默认1 | 仅1允许登录，其他值不允许 |
| register_ip | VARCHAR(45)，默认空串 | 注册时的直连客户端 IP，兼容 IPv6 |
| last_login_at | DATETIME，可空 | 最近登录时间，历史记录未知为空 |
| last_login_ip | VARCHAR(45)，默认空串 | 最近登录的直连客户端 IP |
| login_count | BIGINT UNSIGNED，默认0 | 登录计数，由数据库原子自增 |
| password_changed_at | DATETIME，可空 | 最近设置密码的时间，新注册时记录 |
| created_at | DATETIME，可空 | 创建时间 |
| updated_at | DATETIME，可空 | 更新时间 |
| deleted_at | DATETIME，可空 | 软删除时间，认证查询排除已删除账号 |

手机号和邮箱当前不用于登录，不设置唯一索引，也没有绑定或验证接口。
以后如果将它们作为登录凭据，应先实现验证流程和唯一约束，处理好空值及历史重复数据。
财务余额、积分明细、订单、管理员角色等应使用各自业务表，不塞入用户资料表。

## 注册与登录行为

- 注册接口可选传 `nickname`，最多50字符；空值使用用户名。昵称没有唯一性限制。
- 注册和登录的 IP 由服务器获取直连地址，忽略请求 JSON 中的 client_ip，也不直接信任 X-Forwarded-For。
- 注册时写入注册 IP、密码设置时间，账号初始状态固定为1。
- 登录密码和验证码通过后，记录登录时间/IP及次数；密码错误、验证码错误、停用账号不会更新次数。
- 登录记录写入失败时会撤销刚创建的 Token 并返回错误，不将该 Token 返回给客户端。
- 登录及当前用户响应增加资料与登录记录字段，不返回密码哈希、密码更新时间或注册 IP。
- 头像、手机号、邮箱、性别目前只完成表结构和读取映射；没有新增资料编辑接口。

## 索引

- 主键：`id`。
- 唯一索引：`uk_user_name(name)`，保障并发注册的用户名唯一。
- 联合索引：`idx_user_status_created_at(status, created_at)`，为按状态和创建时间筛选预留。

## 升级方式

完整建表快照位于 `migration/create_user_baseline_20261030080000.go`，来自当前配置数据库的 `SHOW CREATE TABLE user`。
保留真实字段顺序、类型、默认值、注释、索引、字符集和排序规则；不包含用户数据，也不包含当前自增游标。
该快照使用 `utf8mb4_0900_ai_ci`，目标 MySQL 必须支持这一排序规则（当前开发数据库为 MySQL 8）。

用户表只注册 `CreateUserBaseline20261030080000` 这一份完整初始化迁移，原认证、资料两份增量文件已移除。
现有数据库中的旧迁移执行记录保留，不需要删除。
已有数据库补跑快照只会跳过建表并由迁移管理器登记记录，不会将现有表强制改成快照结构。
后续改字段继续新增增量迁移，不要修改已发布快照。
迁移框架另外自动维护 `migrations` 记录表；它不是用户业务表，也无需另写建表迁移。

```powershell
go run ./cmd/migrate -f ./.env
go run ./cmd/make make:dao -f ./.env -t user,migrations
gofmt -w model
go build ./...
go vet ./...
```

这份初始化文件只负责建表，不执行 ALTER、数据回填或账号更新。已存在但结构不同的表需要以后单独写增量迁移。
DAO 生成器现在为 `password_hash` 生成 `json:"-"`，避免意外序列化密码哈希；正式接口仍通过 response 响应结构明确控制字段。

原迁移集成测试文件已清除，当前没有对应测试命令。需要实际验证表结构时，在独立开发或测试库执行并检查，不操作生产数据。
