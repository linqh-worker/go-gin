# 协作与提交规范

开发前阅读 [开发规范](DEVELOPMENT.md)。后续新增模块、接口和任务采用同一分层，不为单个功能另建一套目录或响应格式。

## 一次业务变更的交付范围

按变更需要同时提供：

1. request 参数和校验规则。
2. response 对外字段。
3. controller、logic，以及需要时的模块 DAO。
4. router 注册、认证、权限和数据归属约束。
5. 表结构变更的迁移，以及生成后的 entity/do/dao 底层文件。
6. 说明成功、失败、身份和权限等场景的验证方式；当前不保留测试文件，不自动恢复已删除文件。
7. 新接口或行为变化对应的文档与命令示例。

纯目录移动、注释或低影响修改不要求为实现本身新增形式化测试，但必须完成相应编译和已有检查。

## 版本控制

必须提交：业务源代码、迁移、生成的 Go 文件、go.mod/go.sum、公开配置模板和文档。
不提交：.env、.env.testing 等环境私有配置、日志、临时数据、编译产物或真实账号数据。
密钥按 [加密备份规范](PASSWORD_ENCRYPTION.md) 管理：提交 config/keys/<key_id>/public.pem 和 private.enc.json，不提交 storage/keys 下的明文私钥与备份口令。
当前 .gitignore 已允许提交生成文件，不需要使用 git add -f 绕过忽略规则。

迁移代表可重复的结构变更，不保存用户数据快照；已经执行并共享的迁移不改名、不删除、不重写。当前 user 的完整建表文件是初始化基线，后续变更新增增量文件。
模块名仍为 go-gin；仓库在 GitHub 上叫什么不要求同时改 Go import 路径。若确实修改 module，需一次性更新所有 import 与生成器模板，再测试。

## 分支与提交说明

建议使用 feature/user-profile、fix/auth-token 等短分支名；提交说明描述具体改动，例如：

```text
feat(user): 增加当前用户资料更新
fix(auth): 修复验证码重复使用
docs: 完善迁移和命令行说明
refactor(user): 按业务拆分请求和响应
```

说明正文写清行为变化、数据库变更、验证方式和部署顺序，不用把过程中的每次尝试都写进最终说明。

## 提交前验证

```powershell
gofmt -w app cmd config const cron event internal middleware migration model rest router util
go build ./...
go vet ./...
git status --short
git diff --check
git diff --stat
```

当前没有预置的真实数据库集成测试命令。需要验证业务时只对开发库操作；不要为文档或纯目录改动自动操作数据库。
数据库变更在上线前执行迁移，检查失败输出和迁移记录；代码可编译不代表目标数据库已升级。

## 首次上传 GitHub

先在 GitHub 创建空仓库，下面仓库地址需要替换为自己的实际地址。仓库已有 Git 历史时跳过初始化步骤，保留原有历史。

```powershell
# 仅首次初始化时执行
git init -b main

# 确认本地配置被忽略，公开模板可提交
git check-ignore .env
git status --short --ignored

# 检查将要提交的内容
git add .
git diff --cached --stat
git diff --cached --check
git diff --cached

git commit -m "docs: establish project usage and development conventions"
```

审查暂存内容后，设置实际远端并推送：

```text
git remote add origin https://github.com/你的账号/你的仓库.git
git push -u origin main
```

已有 origin 时先用 `git remote -v` 查看，不要重复添加或随意覆盖。仓库已有提交过的配置不会因加入 .gitignore 自动停止跟踪，需要单独检查暂存内容。
在干净目录重新克隆后，至少验证 `go mod download`、`go build ./...` 和 `go vet ./...`，再按 README 配置数据库启动。
