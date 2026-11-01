package dao

import (
	"context"
	"go-gin/internal/component/db"
	"go-gin/internal/errorx"
	modeldao "go-gin/model/dao"
	"go-gin/model/do"
)

type Menu struct {
	ID, ParentID                            int
	Code, Name, Type, Path, Component, Icon string
	Method, APIPath                         string
	Sort, Status                            int
}

const menuColumns = "id,parent_id,code,name,type,path,component,icon,method,api_path,sort,status"

func Menus(ctx context.Context) ([]Menu, error) {
	rows := []Menu{}
	err := modeldao.SystemMenu.Ctx(ctx).Fields(menuColumns).Order("sort,id").All(&rows)
	return rows, err
}
func RoleMenus(ctx context.Context, id int) ([]int, error) {
	rows := []int{}
	err := modeldao.SystemRoleMenu.Ctx(ctx).Where(do.SystemRoleMenu{RoleId: id}).Order("menu_id").Pluck("menu_id", &rows)
	return rows, err
}
func GrantedMenus(ctx context.Context, id int) ([]int, error) {
	rows := []int{}
	var roleIDs, enabled []int
	if err := modeldao.SystemAdminRole.Ctx(ctx).Where(do.SystemAdminRole{AdminId: id}).Pluck("role_id", &roleIDs); err != nil {
		return rows, err
	}
	if len(roleIDs) == 0 {
		return rows, nil
	}
	if err := modeldao.SystemRole.Ctx(ctx).WhereIn("id", roleIDs).Where(do.SystemRole{Status: 1}).Pluck("id", &enabled); err != nil {
		return rows, err
	}
	if len(enabled) == 0 {
		return rows, nil
	}
	err := modeldao.SystemRoleMenu.Ctx(ctx).WhereIn("role_id", enabled).Distinct().Pluck("menu_id", &rows)
	return rows, err
}

// 所有树变更与菜单授权按同一顺序加锁，避免并发修改产生环或悬空父节点。
func lockMenuTree(ctx context.Context, tx *db.TX) error {
	var role Role
	if err := modeldao.SystemRole.Ctx(ctx).WithTX(tx).Where(do.SystemRole{Code: "super_admin"}).LockUpdate().One(&role); err != nil {
		return err
	}
	if role.ID == 0 {
		return errorx.New(10001, "请先初始化RBAC")
	}
	return nil
}

func SaveMenu(ctx context.Context, m Menu, validate func([]Menu, Menu) error) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		if err := lockMenuTree(ctx, tx); err != nil {
			return err
		}
		var rows []Menu
		if err := modeldao.SystemMenu.Ctx(ctx).WithTX(tx).Fields(menuColumns).Order("sort,id").LockUpdate().All(&rows); err != nil {
			return err
		}
		if err := validate(rows, m); err != nil {
			return err
		}
		if m.ID == 0 {
			_, err := modeldao.SystemMenu.Ctx(ctx).WithTX(tx).Data(menuData(m)).Insert()
			return err
		}
		_, err := modeldao.SystemMenu.Ctx(ctx).WithTX(tx).WherePri(m.ID).Data(menuData(m)).Update()
		return err
	})
}
func DeleteMenu(ctx context.Context, id int) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		if err := lockMenuTree(ctx, tx); err != nil {
			return err
		}
		var m Menu
		if err := modeldao.SystemMenu.Ctx(ctx).WithTX(tx).WherePri(id).One(&m); err != nil {
			return err
		}
		if m.ID == 0 {
			return errorx.New(10001, "菜单不存在")
		}
		count, err := modeldao.SystemMenu.Ctx(ctx).WithTX(tx).Where(do.SystemMenu{ParentId: id}).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return errorx.New(10001, "请先删除子节点")
		}
		if _, err := modeldao.SystemRoleMenu.Ctx(ctx).WithTX(tx).Where(do.SystemRoleMenu{MenuId: id}).Delete(); err != nil {
			return err
		}
		_, err = modeldao.SystemMenu.Ctx(ctx).WithTX(tx).WherePri(id).Delete()
		return err
	})
}
func SetRoleMenus(ctx context.Context, id int, ids []int) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		if err := lockMenuTree(ctx, tx); err != nil {
			return err
		}
		var role Role
		if err := modeldao.SystemRole.Ctx(ctx).WithTX(tx).WherePri(id).LockUpdate().One(&role); err != nil {
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
			if err := modeldao.SystemMenu.Ctx(ctx).WithTX(tx).WherePri(mid).One(&m); err != nil {
				return err
			}
			if m.ID == 0 || m.Status != 1 {
				return errorx.New(10001, "菜单不存在或已停用")
			}
		}
		if _, err := modeldao.SystemRoleMenu.Ctx(ctx).WithTX(tx).Where(do.SystemRoleMenu{RoleId: id}).Delete(); err != nil {
			return err
		}
		for _, mid := range ids {
			if _, err := modeldao.SystemRoleMenu.Ctx(ctx).WithTX(tx).Data(do.SystemRoleMenu{RoleId: id, MenuId: mid}).Insert(); err != nil {
				return err
			}
		}
		return nil
	})
}

func menuData(m Menu) do.SystemMenu {
	return do.SystemMenu{ParentId: m.ParentID, Code: m.Code, Name: m.Name, Type: m.Type, Path: m.Path, Component: m.Component, Icon: m.Icon, Method: m.Method, ApiPath: m.APIPath, Sort: m.Sort, Status: m.Status}
}
