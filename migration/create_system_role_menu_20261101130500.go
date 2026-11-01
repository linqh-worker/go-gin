// system_role_menu 当前结构初始化，只建不存在的表，不修改已有结构或数据。
package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&CreateSystemRoleMenu20261101130500{}) }

type CreateSystemRoleMenu20261101130500 struct{}

func (*CreateSystemRoleMenu20261101130500) Up(db *gorm.DB) error {
	return db.Exec(
		"CREATE TABLE IF NOT EXISTS `system_role_menu` (\n" +
			"  `role_id` int NOT NULL COMMENT '角色ID，关联system_role.id',\n" +
			"  `menu_id` int NOT NULL COMMENT '菜单节点ID，关联system_menu.id',\n" +
			"  PRIMARY KEY (`role_id`,`menu_id`),\n" +
			"  KEY `idx_system_rm_menu` (`menu_id`),\n" +
			"  CONSTRAINT `fk_system_rm_menu` FOREIGN KEY (`menu_id`) REFERENCES `system_menu` (`id`),\n" +
			"  CONSTRAINT `fk_system_rm_role` FOREIGN KEY (`role_id`) REFERENCES `system_role` (`id`)\n" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='后台角色与菜单节点关联'",
	).Error
}
