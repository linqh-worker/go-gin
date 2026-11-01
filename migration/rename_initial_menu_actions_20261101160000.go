package migration

import (
	"go-gin/internal/migration"

	"gorm.io/gorm"
)

func init() { migration.Register(&RenameInitialMenuActions20261101160000{}) }

type RenameInitialMenuActions20261101160000 struct{}

// Rename only legacy placeholders; keep IDs, route bindings and role grants intact.
func (*RenameInitialMenuActions20261101160000) Up(database *gorm.DB) error {
	items := []struct {
		Code, Method, Path, Name string
	}{
		{"endpoint_3", "GET", "/system/admins/:id", "查看管理员详情"},
		{"endpoint_4", "GET", "/system/admins/:id/roles", "查看管理员角色"},
		{"endpoint_7", "GET", "/system/menus", "查看菜单树"},
		{"endpoint_12", "GET", "/system/users/:id", "查看用户详情"},
		{"endpoint_18", "POST", "/system/menus", "新增菜单"},
		{"endpoint_23", "PUT", "/system/menus/:id", "编辑菜单"},
		{"endpoint_24", "PUT", "/system/roles/:id", "编辑角色"},
		{"endpoint_25", "PUT", "/system/roles/:id/menus", "分配角色菜单"},
	}
	return database.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Table("system_menu").
				Where("code = ? AND method = ? AND api_path = ? AND type = ? AND name = ?",
					item.Code, item.Method, item.Path, "api", "系统初始规则").
				UpdateColumn("name", item.Name).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
