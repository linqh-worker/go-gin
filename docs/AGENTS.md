# 项目开发约定

本文件记录仓库内后续代码、配置、测试和文档修改的约定。完整规范见 [开发规范](DEVELOPMENT.md)，命令见 [COMMANDS.md](COMMANDS.md)。

为保持根目录仅有 README，本文件收纳于 docs/；它不再作为仓库根目录的自动加载规则。使用 AI 工具开发时，请先让工具阅读本文件与开发规范。

## 固定约定

- Go module 为 go-gin，最低版本1.26.1。未经任务要求不整体升级依赖或改 module 名。
- HTTP 是单端口模块化应用，API 使用 /api，后台使用 /system。
- 按业务建立 app/<module>/controller/<feature>、logic/<feature>、request/<feature>、response/<feature>。不恢复旧 typing 目录。
- controller 处理绑定、身份提取、业务调用和响应映射；业务规则在 logic；数据库访问在模块 DAO 或共享 DAO。
- logic/dao 接收 context.Context，不导入 Gin、controller 或 router。不用无必要的新层或接口包装简单逻辑。
- 公开接口保持实际入口的 status/msg/data/requestId 外层格式。不要把 httpx 未初始化时的默认字段写成正式接口格式。
- request 字段必须显式定义，客户端不能设置账号状态、权限、密码哈希、当前身份或可信客户端IP。
- 数据库实体不直接作为正式接口响应；通过 response 显式控制输出字段。
- 登录注册密码字段必须经过 RSA-OAEP 解密边界，不增加明文兼容回退。密钥生成、版本加密备份和恢复集中在工具命令，不在业务请求中重新生成密钥。
- API 和 System 的认证独立。新增敏感后台路由前先实现管理员认证与权限，不复用 API Token 作为管理员凭据。
- 查询和写入包含当前用户/资源归属条件，状态、唯一性、金额和重复提交等规则在服务端保证。

## 数据与生成文件

- 新表和字段变更先新增 migration，再执行、生成指定表代码。已执行并共享的迁移不重写。
- 当前 user 初始化迁移只 CREATE TABLE IF NOT EXISTS；不得改成删除、重建或自动清空已有数据。
- 不直接修改 model/entity、model/do、model/dao/internal 或枚举生成文件来实现业务；修改来源后再生成。
- 外层 model/dao/*.go 与 app/<module>/dao 放定制查询。生成器保留外层 DAO，生成文件随迁移提交到仓库。
- 不向仓库写入真实密码、完整 Token、验证码答案、真实业务数据或环境私有配置。同步维护 .env.example 的占位值。
- config/keys/<key_id>/ 的公钥和加密私钥备份允许提交；storage/keys 的明文私钥和备份口令不提交。口令需独立备份。

## 验证与交付

- 修改前检查现有代码，保留用户正在进行的改动，不恢复已删除示例路由。
- 当前测试文件已按用户要求清除，不自动恢复已删除测试文件；交付前运行 go build ./... 和 go vet ./...。
- 后续确需新增测试时按任务范围安排，真实数据库验证仅使用开发库并清理临时数据。
- 新增鉴权、并发写入、资金或迁移行为时提供相应有效验证，明确说明验证范围；不要为纯路径移动制造重复测试。
- 更新受影响的 README、命令、接口和表结构文档。记录实际测试结果；未验证的部署、功能或服务不宣称完成。
