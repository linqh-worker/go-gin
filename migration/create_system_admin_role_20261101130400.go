// system_admin_role 当前结构初始化，只建不存在的表，不修改已有结构或数据。
package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&CreateSystemAdminRole20261101130400{}) }

type CreateSystemAdminRole20261101130400 struct{}

func (*CreateSystemAdminRole20261101130400) Up(db *gorm.DB) error {
	return db.Exec(
		"CREATE TABLE IF NOT EXISTS `system_admin_role` (\n" +
			"  `admin_id` int NOT NULL COMMENT '后台管理员ID，关联system_admin.id',\n" +
			"  `role_id` int NOT NULL COMMENT '角色ID，关联system_role.id',\n" +
			"  PRIMARY KEY (`admin_id`,`role_id`),\n" +
			"  KEY `idx_system_admin_role_role` (`role_id`),\n" +
			"  CONSTRAINT `fk_system_ar_admin` FOREIGN KEY (`admin_id`) REFERENCES `system_admin` (`id`),\n" +
			"  CONSTRAINT `fk_system_ar_role` FOREIGN KEY (`role_id`) REFERENCES `system_role` (`id`)\n" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='后台管理员与角色关联'",
	).Error
}
