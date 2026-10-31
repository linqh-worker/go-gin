// 根据当前配置数据库的 SHOW CREATE TABLE user 导出，仅包含结构，不包含账号数据。
package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&CreateUserBaseline20261030080000{}) }

// 用户表唯一的初始化迁移，包含账号认证、资料和登录记录完整字段。
// 已存在的 user 表不会被重建，也不会在此迁移中修改结构或数据。
type CreateUserBaseline20261030080000 struct{}

func (m *CreateUserBaseline20261030080000) Up(db *gorm.DB) error {
	return db.Exec(userBaselineDDL).Error
}

const userBaselineDDL = "CREATE TABLE IF NOT EXISTS `user` (\n" +
	"  `id` int NOT NULL AUTO_INCREMENT,\n" +
	"  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户名',\n" +
	"  `age` int DEFAULT NULL COMMENT '年龄',\n" +
	"  `status` int DEFAULT '1' COMMENT '状态',\n" +
	"  `password_hash` varchar(100) DEFAULT NULL COMMENT 'bcrypt密码哈希，旧账号为空',\n" +
	"  `nickname` varchar(50) NOT NULL DEFAULT '' COMMENT '昵称',\n" +
	"  `avatar_url` varchar(255) NOT NULL DEFAULT '' COMMENT '头像地址',\n" +
	"  `phone` varchar(32) DEFAULT NULL COMMENT '手机号，保存资料不代表已验证',\n" +
	"  `email` varchar(254) DEFAULT NULL COMMENT '邮箱，保存资料不代表已验证',\n" +
	"  `gender` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '性别：0未知，1男，2女',\n" +
	"  `register_ip` varchar(45) NOT NULL DEFAULT '' COMMENT '注册时直连客户端IP',\n" +
	"  `last_login_at` datetime DEFAULT NULL COMMENT '最近成功登录时间',\n" +
	"  `last_login_ip` varchar(45) NOT NULL DEFAULT '' COMMENT '最近成功登录IP',\n" +
	"  `login_count` bigint unsigned NOT NULL DEFAULT '0' COMMENT '成功登录次数',\n" +
	"  `password_changed_at` datetime DEFAULT NULL COMMENT '密码最后设置时间，未知为空',\n" +
	"  `created_at` datetime DEFAULT NULL COMMENT '创建时间',\n" +
	"  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',\n" +
	"  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  UNIQUE KEY `uk_user_name` (`name`),\n" +
	"  KEY `idx_user_status_created_at` (`status`,`created_at`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户端账号与资料'"
