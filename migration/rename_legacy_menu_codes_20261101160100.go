package migration

import (
	"fmt"
	"go-gin/internal/migration"

	"gorm.io/gorm"
)

func init() { migration.Register(&RenameLegacyMenuCodes20261101160100{}) }

type RenameLegacyMenuCodes20261101160100 struct{}

// Replace legacy placeholders while preserving menu IDs and role associations.
func (*RenameLegacyMenuCodes20261101160100) Up(database *gorm.DB) error {
	items := []struct{ Old, Code, Method, Path string }{
		{"endpoint_3", "admin_detail", "GET", "/system/admins/:id"},
		{"endpoint_4", "admin_roles_read", "GET", "/system/admins/:id/roles"},
		{"endpoint_7", "menu_read", "GET", "/system/menus"},
		{"endpoint_12", "user_detail", "GET", "/system/users/:id"},
		{"endpoint_18", "menu_create", "POST", "/system/menus"},
		{"endpoint_23", "menu_update", "PUT", "/system/menus/:id"},
		{"endpoint_24", "role_update", "PUT", "/system/roles/:id"},
		{"endpoint_25", "role_menu_assign", "PUT", "/system/roles/:id/menus"},
	}
	return database.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Table("system_menu").
				Where("code = ? AND method = ? AND api_path = ? AND type = ?",
					item.Old, item.Method, item.Path, "api").
				UpdateColumn("code", item.Code).Error; err != nil {
				return fmt.Errorf("更新菜单标识 %s -> %s 失败: %w", item.Old, item.Code, err)
			}
		}
		return nil
	})
}
