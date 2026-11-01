// system_admin 当前结构初始化，只建不存在的表，不修改已有结构或数据。
package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&CreateSystemAdmin20261101130100{}) }

type CreateSystemAdmin20261101130100 struct{}

func (*CreateSystemAdmin20261101130100) Up(db *gorm.DB) error {
	return db.Exec(
		"CREATE TABLE IF NOT EXISTS `system_admin` (\n" +
			"  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n" +
			"  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '登录用户名，全表唯一',\n" +
			"  `age` int NOT NULL DEFAULT '0' COMMENT '年龄，0表示未填写',\n" +
			"  `status` int NOT NULL DEFAULT '1' COMMENT '状态：0停用，1启用',\n" +
			"  `password_hash` varchar(100) DEFAULT NULL COMMENT 'bcrypt密码哈希，旧账号未设置时为NULL',\n" +
			"  `nickname` varchar(50) NOT NULL DEFAULT '' COMMENT '昵称',\n" +
			"  `avatar_url` varchar(255) NOT NULL DEFAULT '' COMMENT '头像地址',\n" +
			"  `phone` varchar(32) NOT NULL DEFAULT '' COMMENT '手机号资料，不代表已验证',\n" +
			"  `email` varchar(254) NOT NULL DEFAULT '' COMMENT '邮箱资料，不代表已验证',\n" +
			"  `gender` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '性别：0未知，1男，2女',\n" +
			"  `register_ip` varchar(45) NOT NULL DEFAULT '' COMMENT '注册时直连客户端IP',\n" +
			"  `last_login_at` datetime DEFAULT NULL COMMENT '最近成功登录时间，未登录为NULL',\n" +
			"  `last_login_ip` varchar(45) NOT NULL DEFAULT '' COMMENT '最近成功登录的直连客户端IP',\n" +
			"  `login_count` bigint unsigned NOT NULL DEFAULT '0' COMMENT '成功登录次数',\n" +
			"  `password_changed_at` datetime DEFAULT NULL COMMENT '最近设置密码时间，未知为NULL',\n" +
			"  `created_at` datetime DEFAULT NULL COMMENT '创建时间',\n" +
			"  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',\n" +
			"  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',\n" +
			"  PRIMARY KEY (`id`),\n" +
			"  UNIQUE KEY `uk_system_admin_name` (`name`),\n" +
			"  KEY `idx_system_admin_status_created_at` (`status`,`created_at`)\n" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='后台管理员账号与资料'",
	).Error
}
