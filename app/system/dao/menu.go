package dao

import (
	"context"
	"go-gin/internal/component/db"
	"go-gin/internal/errorx"
)

type Menu struct {
	ID, ParentID                                            int
	Code, Name, Type, Path, Component, Icon, PermissionCode string
	Sort, Status                                            int
}

const menuColumns = "id,parent_id,code,name,type,path,component,icon,permission_code,sort,status"

func Menus(ctx context.Context) ([]Menu, error) {
	rows := []Menu{}
	err := db.GetScan(ctx, &rows, "SELECT "+menuColumns+" FROM system_menu ORDER BY sort,id")
	return rows, err
}
func RoleMenus(ctx context.Context, id int) ([]int, error) {
	rows := []int{}
	err := db.GetScan(ctx, &rows, "SELECT menu_id FROM system_role_menu WHERE role_id=? ORDER BY menu_id", id)
	return rows, err
}
func GrantedMenus(ctx context.Context, id int) ([]int, error) {
	rows := []int{}
	err := db.GetScan(ctx, &rows, `SELECT DISTINCT rm.menu_id FROM system_role_menu rm JOIN system_role r ON r.id=rm.role_id JOIN system_admin_role ar ON ar.role_id=r.id WHERE ar.admin_id=? AND r.status=1`, id)
	return rows, err
}

// 所有树变更与菜单授权按同一顺序加锁，避免并发修改产生环或悬空父节点。
func lockMenuTree(tx *db.TX) error {
	var role Role
	if err := tx.Raw("SELECT id FROM system_role WHERE code='super_admin' FOR UPDATE").Scan(&role).Error(); err != nil {
		return err
	}
	if role.ID == 0 {
		return errorx.New(10001, "请先初始化RBAC")
	}
	return nil
}

func SaveMenu(ctx context.Context, m Menu, validate func([]Menu, Menu) error) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		if err := lockMenuTree(tx); err != nil {
			return err
		}
		var rows []Menu
		if err := tx.Raw("SELECT " + menuColumns + " FROM system_menu ORDER BY sort,id FOR UPDATE").Scan(&rows).Error(); err != nil {
			return err
		}
		if err := validate(rows, m); err != nil {
			return err
		}
		if m.PermissionCode != "" {
			var p Permission
			if err := tx.Raw("SELECT id FROM system_permission WHERE code=?", m.PermissionCode).Scan(&p).Error(); err != nil {
				return err
			}
			if p.ID == 0 {
				return errorx.New(10001, "关联接口权限不存在")
			}
		}
		if m.ID == 0 {
			_, err := tx.Exec("INSERT INTO system_menu(parent_id,code,name,type,path,component,icon,permission_code,sort,status) VALUES(?,?,?,?,?,?,?,?,?,?)", m.ParentID, m.Code, m.Name, m.Type, m.Path, m.Component, m.Icon, m.PermissionCode, m.Sort, m.Status)
			return err
		}
		_, err := tx.Exec("UPDATE system_menu SET parent_id=?,code=?,name=?,type=?,path=?,component=?,icon=?,permission_code=?,sort=?,status=? WHERE id=?", m.ParentID, m.Code, m.Name, m.Type, m.Path, m.Component, m.Icon, m.PermissionCode, m.Sort, m.Status, m.ID)
		return err
	})
}
func DeleteMenu(ctx context.Context, id int) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		if err := lockMenuTree(tx); err != nil {
			return err
		}
		var m Menu
		if err := tx.Raw("SELECT id FROM system_menu WHERE id=?", id).Scan(&m).Error(); err != nil {
			return err
		}
		if m.ID == 0 {
			return errorx.New(10001, "菜单不存在")
		}
		var row struct{ N int }
		if err := tx.Raw("SELECT COUNT(*) n FROM system_menu WHERE parent_id=?", id).Scan(&row).Error(); err != nil {
			return err
		}
		if row.N > 0 {
			return errorx.New(10001, "请先删除子节点")
		}
		if _, err := tx.Exec("DELETE FROM system_role_menu WHERE menu_id=?", id); err != nil {
			return err
		}
		_, err := tx.Exec("DELETE FROM system_menu WHERE id=?", id)
		return err
	})
}
func SetRoleMenus(ctx context.Context, id int, ids []int) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		if err := lockMenuTree(tx); err != nil {
			return err
		}
		var role Role
		if err := tx.Raw("SELECT id,code FROM system_role WHERE id=? FOR UPDATE", id).Scan(&role).Error(); err != nil {
			return err
		}
		if role.ID == 0 || role.Code == "super_admin" {
			return errorx.New(10001, "角色不存在或为不可修改的超级管理员角色")
		}
		seen := map[int]bool{}
		for _, mid := range ids {
			if mid <= 0 || seen[mid] {
				return errorx.New(10001, "菜单ID无效或重复")
			}
			seen[mid] = true
			var m Menu
			if err := tx.Raw("SELECT id,status FROM system_menu WHERE id=?", mid).Scan(&m).Error(); err != nil {
				return err
			}
			if m.ID == 0 || m.Status != 1 {
				return errorx.New(10001, "菜单不存在或已停用")
			}
		}
		if _, err := tx.Exec("DELETE FROM system_role_menu WHERE role_id=?", id); err != nil {
			return err
		}
		for _, mid := range ids {
			if _, err := tx.Exec("INSERT INTO system_role_menu(role_id,menu_id) VALUES(?,?)", id, mid); err != nil {
				return err
			}
		}
		return nil
	})
}
