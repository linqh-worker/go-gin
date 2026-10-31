package dao

import (
	"context"
	"go-gin/internal/component/db"
	"go-gin/internal/errorx"
)

type Role struct {
	ID     int    `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}
type Permission struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
type Access struct {
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
type AdminSummary struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Status   int    `json:"status"`
}

func Admins(ctx context.Context) ([]AdminSummary, error) {
	rows := []AdminSummary{}
	err := db.GetScan(ctx, &rows, "SELECT id,name AS username,nickname,status FROM system_admin WHERE deleted_at IS NULL ORDER BY id LIMIT 100")
	return rows, err
}

// LoadAccess 每次从库读取，角色停用、撤权无需等待 Token 过期。
func LoadAccess(ctx context.Context, id int) (Access, error) {
	a := Access{Roles: []string{}, Permissions: []string{}}
	var roles []Role
	err := db.GetScan(ctx, &roles, `SELECT r.id,r.code,r.name,r.status FROM system_role r JOIN system_admin_role ar ON ar.role_id=r.id JOIN system_admin a ON a.id=ar.admin_id WHERE a.id=? AND a.status=1 AND a.deleted_at IS NULL AND r.status=1 ORDER BY r.id`, id)
	if err != nil {
		return a, err
	}
	super := false
	for _, r := range roles {
		a.Roles = append(a.Roles, r.Code)
		if r.Code == "super_admin" {
			super = true
		}
	}
	if super {
		a.Permissions = []string{"*"}
		return a, nil
	}
	var permissions []Permission
	err = db.GetScan(ctx, &permissions, `SELECT DISTINCT p.id,p.code,p.name FROM system_permission p JOIN system_role_permission rp ON rp.permission_id=p.id JOIN system_role r ON r.id=rp.role_id JOIN system_admin_role ar ON ar.role_id=r.id WHERE ar.admin_id=? AND r.status=1 ORDER BY p.id`, id)
	for _, p := range permissions {
		a.Permissions = append(a.Permissions, p.Code)
	}
	return a, err
}
func Roles(ctx context.Context) ([]Role, error) {
	rows := []Role{}
	err := db.GetScan(ctx, &rows, "SELECT id,code,name,status FROM system_role ORDER BY id")
	return rows, err
}
func Permissions(ctx context.Context) ([]Permission, error) {
	rows := []Permission{}
	err := db.GetScan(ctx, &rows, "SELECT id,code,name FROM system_permission ORDER BY id")
	return rows, err
}
func RolePermissions(ctx context.Context, id int) ([]int, error) {
	rows := []int{}
	err := db.GetScan(ctx, &rows, "SELECT permission_id FROM system_role_permission WHERE role_id=? ORDER BY permission_id", id)
	return rows, err
}
func AdminRoles(ctx context.Context, id int) ([]int, error) {
	rows := []int{}
	err := db.GetScan(ctx, &rows, "SELECT role_id FROM system_admin_role WHERE admin_id=? ORDER BY role_id", id)
	return rows, err
}

func SaveRole(ctx context.Context, id int, code, name string, status int, permissionIDs []int) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		if id == 0 {
			if _, err := tx.Exec("INSERT INTO system_role(code,name,status) VALUES(?,?,?)", code, name, status); err != nil {
				return err
			}
			var row struct{ ID int }
			if err := tx.Raw("SELECT id FROM system_role WHERE code=? FOR UPDATE", code).Scan(&row).Error(); err != nil {
				return err
			}
			id = row.ID
		} else {
			var row Role
			if err := tx.Raw("SELECT id,code FROM system_role WHERE id=? FOR UPDATE", id).Scan(&row).Error(); err != nil {
				return err
			}
			if row.ID == 0 {
				return errorx.New(10001, "角色不存在")
			}
			if row.Code == "super_admin" {
				return errorx.New(403, "超级管理员角色不允许通过接口修改")
			}
			if _, err := tx.Exec("UPDATE system_role SET name=?,status=? WHERE id=?", name, status, id); err != nil {
				return err
			}
		}
		if _, err := tx.Exec("DELETE FROM system_role_permission WHERE role_id=?", id); err != nil {
			return err
		}
		seen := map[int]bool{}
		for _, pid := range permissionIDs {
			if pid <= 0 || seen[pid] {
				return errorx.New(10001, "权限ID无效或重复")
			}
			seen[pid] = true
			var p Permission
			if err := tx.Raw("SELECT id FROM system_permission WHERE id=? FOR UPDATE", pid).Scan(&p).Error(); err != nil {
				return err
			}
			if p.ID == 0 {
				return errorx.New(10001, "权限不存在")
			}
			if _, err := tx.Exec("INSERT INTO system_role_permission(role_id,permission_id) VALUES(?,?)", id, pid); err != nil {
				return err
			}
		}
		return nil
	})
}

func SetAdminRoles(ctx context.Context, adminID int, ids []int) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		var admin struct{ ID int }
		if err := tx.Raw("SELECT id FROM system_admin WHERE id=? AND deleted_at IS NULL FOR UPDATE", adminID).Scan(&admin).Error(); err != nil {
			return err
		}
		if admin.ID == 0 {
			return errorx.New(10001, "管理员不存在")
		}
		var super struct{ N int }
		if err := tx.Raw("SELECT COUNT(*) n FROM system_admin_role ar JOIN system_role r ON r.id=ar.role_id WHERE ar.admin_id=? AND r.code='super_admin'", adminID).Scan(&super).Error(); err != nil {
			return err
		}
		if super.N > 0 {
			return errorx.New(403, "不能通过接口修改超级管理员的角色")
		}
		for _, id := range ids {
			var r Role
			if err := tx.Raw("SELECT id,code,status FROM system_role WHERE id=? FOR UPDATE", id).Scan(&r).Error(); err != nil {
				return err
			}
			if r.ID == 0 || r.Status != 1 {
				return errorx.New(10001, "角色不存在或已停用")
			}
			if r.Code == "super_admin" {
				return errorx.New(403, "超级管理员只能通过服务器命令授予")
			}
		}
		if _, err := tx.Exec("DELETE FROM system_admin_role WHERE admin_id=?", adminID); err != nil {
			return err
		}
		seen := map[int]bool{}
		for _, id := range ids {
			if seen[id] {
				return errorx.New(10001, "角色ID重复")
			}
			seen[id] = true
			if _, err := tx.Exec("INSERT INTO system_admin_role(admin_id,role_id) VALUES(?,?)", adminID, id); err != nil {
				return err
			}
		}
		return nil
	})
}
