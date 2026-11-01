package migration

import (
	"go-gin/internal/migration"

	"gorm.io/gorm"
)

func init() { migration.Register(&NormalizeMenuButtonCodes20261101160200{}) }

type NormalizeMenuButtonCodes20261101160200 struct{}

// Make legacy and fresh installs use the same stable frontend button identifiers.
func (*NormalizeMenuButtonCodes20261101160200) Up(database *gorm.DB) error {
	items := []struct{ Old, Code, Method, Path string }{
		{"role_manage", "role_create", "POST", "/system/roles"},
		{"role_menus", "role_menus_read", "GET", "/system/roles/:id/menus"},
		{"menu_manage", "menu_delete", "DELETE", "/system/menus/:id"},
		{"role_menus", "role_menu_assign", "PUT", "/system/roles/:id/menus"},
	}
	return database.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Table("system_menu").
				Where("code = ? AND method = ? AND api_path = ?", item.Old, item.Method, item.Path).
				UpdateColumn("code", item.Code).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
