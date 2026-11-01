package dao

import (
	"context"
	"go-gin/internal/component/db"
	"go-gin/internal/errorx"
	modeldao "go-gin/model/dao"
	"go-gin/model/do"
)

type Role struct {
	ID     int    `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}
type Access struct {
	Roles   []string `json:"roles"`
	MenuIDs []int    `json:"menu_ids"`
}
type AdminSummary struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Status   int    `json:"status"`
}

func Admins(ctx context.Context) ([]AdminSummary, error) {
	rows := []AdminSummary{}
	err := modeldao.SystemAdmin.Ctx(ctx).Fields("id,name AS username,nickname,status").WhereNull("deleted_at").Order("id").Limit(100).All(&rows)
	return rows, err
}
func LoadAccess(ctx context.Context, id int) (Access, error) {
	a := Access{Roles: []string{}, MenuIDs: []int{}}
	exists, err := modeldao.SystemAdmin.Ctx(ctx).WherePri(id).Where(do.SystemAdmin{Status: 1}).WhereNull("deleted_at").Exist()
	if err != nil || !exists {
		return a, err
	}
	ids, err := AdminRoles(ctx, id)
	if err != nil || len(ids) == 0 {
		return a, err
	}
	var roles []Role
	if err = modeldao.SystemRole.Ctx(ctx).WhereIn("id", ids).Where(do.SystemRole{Status: 1}).Order("id").All(&roles); err != nil {
		return a, err
	}
	for _, r := range roles {
		a.Roles = append(a.Roles, r.Code)
	}
	a.MenuIDs, err = GrantedMenus(ctx, id)
	return a, err
}
func Roles(ctx context.Context) ([]Role, error) {
	rows := []Role{}
	err := modeldao.SystemRole.Ctx(ctx).Order("id").All(&rows)
	return rows, err
}
func AdminRoles(ctx context.Context, id int) ([]int, error) {
	rows := []int{}
	err := modeldao.SystemAdminRole.Ctx(ctx).Where(do.SystemAdminRole{AdminId: id}).Order("role_id").Pluck("role_id", &rows)
	return rows, err
}
func SaveRole(ctx context.Context, id int, code, name string, status int) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		q := modeldao.SystemRole.Ctx(ctx).WithTX(tx)
		if id == 0 {
			_, err := q.Data(do.SystemRole{Code: code, Name: name, Status: status}).Insert()
			return err
		}
		var row Role
		if err := q.WherePri(id).LockUpdate().One(&row); err != nil {
			return err
		}
		if row.ID == 0 || row.Code == "super_admin" {
			return errorx.New(10001, "角色不存在或为不可修改的超级管理员角色")
		}
		_, err := q.WherePri(id).Data(do.SystemRole{Name: name, Status: status}).Update()
		return err
	})
}
func hasSuperRole(ctx context.Context, tx *db.TX, adminID int) (bool, error) {
	var role Role
	if err := modeldao.SystemRole.Ctx(ctx).WithTX(tx).Where(do.SystemRole{Code: "super_admin"}).One(&role); err != nil {
		return false, err
	}
	if role.ID == 0 {
		return false, nil
	}
	return modeldao.SystemAdminRole.Ctx(ctx).WithTX(tx).Where(do.SystemAdminRole{AdminId: adminID, RoleId: role.ID}).Exist()
}
func SetAdminRoles(ctx context.Context, adminID int, ids []int) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		var admin struct {
			ID          int
			IsProtected bool
		}
		if err := modeldao.SystemAdmin.Ctx(ctx).WithTX(tx).WherePri(adminID).WhereNull("deleted_at").LockUpdate().One(&admin); err != nil {
			return err
		}
		if admin.ID == 0 {
			return errorx.New(10001, "管理员不存在")
		}
		if admin.IsProtected {
			return errorx.New(403, "演示管理员受保护，不允许修改角色")
		}
		super, err := hasSuperRole(ctx, tx, adminID)
		if err != nil {
			return err
		}
		if super {
			return errorx.New(403, "不能通过接口修改超级管理员的角色")
		}
		seen := map[int]bool{}
		for _, id := range ids {
			if id <= 0 || seen[id] {
				return errorx.New(10001, "角色ID无效或重复")
			}
			seen[id] = true
			var r Role
			if err := modeldao.SystemRole.Ctx(ctx).WithTX(tx).WherePri(id).LockUpdate().One(&r); err != nil {
				return err
			}
			if r.ID == 0 || r.Status != 1 {
				return errorx.New(10001, "角色不存在或已停用")
			}
			if r.Code == "super_admin" {
				return errorx.New(403, "超级管理员只能通过服务器命令授予")
			}
		}
		q := modeldao.SystemAdminRole.Ctx(ctx).WithTX(tx)
		if _, err := q.Where(do.SystemAdminRole{AdminId: adminID}).Delete(); err != nil {
			return err
		}
		for _, id := range ids {
			if _, err := q.Data(do.SystemAdminRole{AdminId: adminID, RoleId: id}).Insert(); err != nil {
				return err
			}
		}
		return nil
	})
}
