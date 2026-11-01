package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&CompleteAdminManagement20261101170000{}) }

type CompleteAdminManagement20261101170000 struct{}

func (*CompleteAdminManagement20261101170000) Up(database *gorm.DB) error {
	if !database.Migrator().HasColumn("system_admin", "is_protected") {
		if err := database.Exec("ALTER TABLE system_admin ADD COLUMN is_protected tinyint unsigned NOT NULL DEFAULT 0 COMMENT '演示账号保护：1禁止通过HTTP修改资料、密码、状态及角色'").Error; err != nil {
			return err
		}
	}
	return database.Transaction(func(tx *gorm.DB) error {
		icons := map[string]string{"system_root": "ri:settings-3-line", "user_root": "ri:group-line", "admin_list": "ri:admin-line", "role_list": "ri:shield-user-line", "menu_tree": "ri:menu-line", "user_list": "ri:user-3-line"}
		for code, icon := range icons {
			if err := tx.Table("system_menu").Where("code = ? AND (icon = '' OR icon IS NULL)", code).UpdateColumn("icon", icon).Error; err != nil {
				return err
			}
		}
		return tx.Exec("INSERT INTO system_menu(parent_id,code,name,type,method,api_path,sort,status) SELECT id, 'admin_password', '修改管理员密码', 'button', 'PUT', '/system/admins/:id/password', 80, 1 FROM system_menu WHERE code='admin_list' AND NOT EXISTS (SELECT 1 FROM (SELECT code FROM system_menu) AS existing WHERE code='admin_password')").Error
	})
}
