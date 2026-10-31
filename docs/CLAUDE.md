# 项目说明

本仓库统一开发约定见 [AGENTS.md](AGENTS.md)，完整业务规范见 [开发规范](DEVELOPMENT.md)。本文件收纳于 docs/，不作为根目录自动加载配置。

请先阅读 [README.md](../README.md) 与 [MODULES.md](MODULES.md)，使用 [COMMANDS.md](COMMANDS.md) 中的真实命令。
Go 版本为1.26.1或更高。请求、响应已拆成 request/response 并按业务分包，不使用旧 typing 结构。
正式接口外层字段为 code/msg/data/requestId，API 认证使用 Redis 随机 Token，不是 JWT；System 管理员认证尚未实现。

不另行维护与 AGENTS.md 冲突的目录、命令或响应格式约定。
