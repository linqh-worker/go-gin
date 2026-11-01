package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&SeedSystem20261101130600{}) }

type SeedSystem20261101130600 struct{}
type initialMenu struct {
	Code, Name, Parent, Type, Path, Component, Method, APIPath string
	Sort                                                       int
}

// 只插入缺少的基础节点及新角色；不覆盖已有菜单配置或重新授予已有角色权限。
func (*SeedSystem20261101130600) Up(database *gorm.DB) error {
	return database.Transaction(func(db *gorm.DB) error {
		nodes := []initialMenu{
			{Code: "system_root", Name: "系统管理", Type: "directory", Sort: 10},
			{Code: "user_root", Name: "用户管理", Type: "directory", Sort: 20},
			{Code: "admin_list", Name: "管理员管理", Parent: "system_root", Type: "menu", Path: "/system/admins", Component: "system/admin/index", Sort: 10},
			{Code: "role_list", Name: "角色管理", Parent: "system_root", Type: "menu", Path: "/system/roles", Component: "system/role/index", Sort: 20},
			{Code: "menu_tree", Name: "菜单管理", Parent: "system_root", Type: "menu", Path: "/system/menus", Component: "system/menu/index", Sort: 30},
			{Code: "user_list", Name: "用户列表", Parent: "user_root", Type: "menu", Path: "/system/users", Component: "system/user/index", Sort: 10},
			{Code: "admin_read", Name: "查看管理员", Parent: "admin_list", Type: "button", Method: "GET", APIPath: "/system/admins", Sort: 10},
			{Code: "admin_detail", Name: "管理员详情", Parent: "admin_list", Type: "api", Method: "GET", APIPath: "/system/admins/:id", Sort: 20},
			{Code: "admin_create", Name: "新增管理员", Parent: "admin_list", Type: "button", Method: "POST", APIPath: "/system/admins", Sort: 30},
			{Code: "admin_update", Name: "编辑管理员", Parent: "admin_list", Type: "button", Method: "PATCH", APIPath: "/system/admins/:id", Sort: 40},
			{Code: "admin_status", Name: "启停管理员", Parent: "admin_list", Type: "button", Method: "PUT", APIPath: "/system/admins/:id/status", Sort: 50},
			{Code: "admin_roles_read", Name: "查看管理员角色", Parent: "admin_list", Type: "api", Method: "GET", APIPath: "/system/admins/:id/roles", Sort: 60},
			{Code: "admin_roles", Name: "分配角色", Parent: "admin_list", Type: "button", Method: "PUT", APIPath: "/system/admins/:id/roles", Sort: 70},
			{Code: "role_read", Name: "查看角色", Parent: "role_list", Type: "button", Method: "GET", APIPath: "/system/roles", Sort: 10},
			{Code: "role_create", Name: "新增角色", Parent: "role_list", Type: "button", Method: "POST", APIPath: "/system/roles", Sort: 20},
			{Code: "role_update", Name: "编辑角色", Parent: "role_list", Type: "button", Method: "PUT", APIPath: "/system/roles/:id", Sort: 30},
			{Code: "role_menus_read", Name: "查看角色菜单", Parent: "role_list", Type: "api", Method: "GET", APIPath: "/system/roles/:id/menus", Sort: 40},
			{Code: "role_menus", Name: "勾选角色菜单", Parent: "role_list", Type: "button", Method: "PUT", APIPath: "/system/roles/:id/menus", Sort: 50},
			{Code: "menu_read", Name: "查看菜单树", Parent: "menu_tree", Type: "api", Method: "GET", APIPath: "/system/menus", Sort: 10},
			{Code: "menu_create", Name: "新增菜单节点", Parent: "menu_tree", Type: "button", Method: "POST", APIPath: "/system/menus", Sort: 20},
			{Code: "menu_update", Name: "编辑菜单节点", Parent: "menu_tree", Type: "button", Method: "PUT", APIPath: "/system/menus/:id", Sort: 30},
			{Code: "menu_delete", Name: "删除菜单节点", Parent: "menu_tree", Type: "button", Method: "DELETE", APIPath: "/system/menus/:id", Sort: 40},
			{Code: "user_read", Name: "查看用户", Parent: "user_list", Type: "button", Method: "GET", APIPath: "/system/users", Sort: 10},
			{Code: "user_detail", Name: "用户详情", Parent: "user_list", Type: "api", Method: "GET", APIPath: "/system/users/:id", Sort: 20},
			{Code: "user_create", Name: "新增用户", Parent: "user_list", Type: "button", Method: "POST", APIPath: "/system/users", Sort: 30},
			{Code: "user_update", Name: "编辑用户", Parent: "user_list", Type: "button", Method: "PATCH", APIPath: "/system/users/:id", Sort: 40},
			{Code: "user_status", Name: "启停用户", Parent: "user_list", Type: "button", Method: "PUT", APIPath: "/system/users/:id/status", Sort: 50},
		}
		ids := map[string]int{}
		for _, node := range nodes {
			var existing struct{ ID int }
			if err := db.Raw("SELECT id FROM system_menu WHERE code=?", node.Code).Scan(&existing).Error; err != nil {
				return err
			}
			// 现有库从旧版转换过来时节点编码可能不同，用方法+路径复用原节点，避免重复创建。
			if existing.ID == 0 && node.Method != "" {
				if err := db.Raw("SELECT id FROM system_menu WHERE method=? AND api_path=? ORDER BY id LIMIT 1", node.Method, node.APIPath).Scan(&existing).Error; err != nil {
					return err
				}
			}
			if existing.ID == 0 {
				if err := db.Exec("INSERT INTO system_menu(parent_id,code,name,type,path,component,method,api_path,sort,status) VALUES(?,?,?,?,?,?,?,?,?,1)", ids[node.Parent], node.Code, node.Name, node.Type, node.Path, node.Component, node.Method, node.APIPath, node.Sort).Error; err != nil {
					return err
				}
				if err := db.Raw("SELECT id FROM system_menu WHERE code=?", node.Code).Scan(&existing).Error; err != nil {
					return err
				}
			}
			ids[node.Code] = existing.ID
		}
		roles := []struct {
			Code, Name string
			Menus      []string
		}{
			{"super_admin", "超级管理员", nil},
			{"auditor", "只读审核员", []string{"system_root", "admin_list", "admin_read", "admin_detail", "admin_roles_read", "role_list", "role_read", "user_root", "user_list", "user_read", "user_detail"}},
			{"user_operator", "用户运营员", []string{"user_root", "user_list", "user_read", "user_detail", "user_create", "user_update", "user_status"}},
		}
		for _, role := range roles {
			var existing struct{ ID int }
			if err := db.Raw("SELECT id FROM system_role WHERE code=?", role.Code).Scan(&existing).Error; err != nil {
				return err
			}
			if existing.ID != 0 {
				continue
			}
			if err := db.Exec("INSERT INTO system_role(code,name,status) VALUES(?,?,1)", role.Code, role.Name).Error; err != nil {
				return err
			}
			if err := db.Raw("SELECT id FROM system_role WHERE code=?", role.Code).Scan(&existing).Error; err != nil {
				return err
			}
			for _, code := range role.Menus {
				if err := db.Exec("INSERT INTO system_role_menu(role_id,menu_id) VALUES(?,?)", existing.ID, ids[code]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
