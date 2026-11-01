// system_role 当前结构初始化，只建不存在的表，不修改已有结构或数据。
package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&CreateSystemRole20261101130200{}) }

type CreateSystemRole20261101130200 struct{}

func (*CreateSystemRole20261101130200) Up(db *gorm.DB) error {
	return db.Exec(
		"CREATE TABLE IF NOT EXISTS `system_role` (\n" +
			"  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n" +
			"  `code` varchar(64) NOT NULL COMMENT '唯一业务编码',\n" +
			"  `name` varchar(100) NOT NULL COMMENT '名称',\n" +
			"  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：0停用，1启用',\n" +
			"  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n" +
			"  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n" +
			"  PRIMARY KEY (`id`),\n" +
			"  UNIQUE KEY `uk_system_role_code` (`code`)\n" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='后台角色'",
	).Error
}
