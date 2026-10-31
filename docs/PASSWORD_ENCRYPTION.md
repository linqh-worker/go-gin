# 请求体密码加密已停用

当前登录注册直接提交密码字符串，不需要 encryption，也不再提供 /api/auth/public-key。数据库继续使用 bcrypt 哈希。接口示例见 [API_AUTH.md](API_AUTH.md)。

历史密钥工具与文件保留，避免丢失已有备份；它们不参与当前 HTTP 启动和认证。
