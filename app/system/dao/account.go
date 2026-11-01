package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"go-gin/internal/component/db"
	"go-gin/internal/errorx"
	modeldao "go-gin/model/dao"
	"time"
)

func accountModel(ctx context.Context, admin bool) *db.Model {
	if admin {
		return modeldao.SystemAdmin.Ctx(ctx)
	}
	return modeldao.User.Ctx(ctx)
}

const profileColumns = "id,name,nickname,avatar_url,phone,email,age,gender,status,login_count,last_login_at,created_at,updated_at"

func ListAccounts(ctx context.Context, admin bool, page, size int, keyword string, status *int) ([]AuthUser, int64, error) {
	query := accountModel(ctx, admin).WhereNull("deleted_at")
	if keyword != "" {
		query = query.Where("(name LIKE ? OR nickname LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}
	rows := []AuthUser{}
	err = query.Fields(accountProfileColumns(admin)).Order("id DESC").Page(page, size).All(&rows)
	return rows, total, err
}
func Account(ctx context.Context, admin bool, id int) (*AuthUser, error) {
	var row AuthUser
	err := accountModel(ctx, admin).Fields(accountProfileColumns(admin)).WherePri(id).WhereNull("deleted_at").One(&row)
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, errorx.New(10001, "账号不存在")
	}
	return &row, nil
}
func CreateAccount(ctx context.Context, admin bool, user *AuthUser) error {
	id, err := accountModel(ctx, admin).Data(authData(user)).InsertAndGetId()
	if err == nil {
		user.ID = int(id)
	}
	var duplicate *mysql.MySQLError
	if errors.As(err, &duplicate) && duplicate.Number == 1062 {
		return errorx.New(20007, "用户名已存在")
	}
	return err
}
func UpdateAccount(ctx context.Context, admin bool, id int, fields map[string]any, statusChange bool) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		var row struct {
			ID          int
			IsProtected bool
		}
		if err := accountModel(ctx, admin).WithTX(tx).WherePri(id).WhereNull("deleted_at").LockUpdate().One(&row); err != nil {
			return err
		}
		if row.ID == 0 {
			return errorx.New(10001, "账号不存在")
		}
		if admin && row.IsProtected {
			return errorx.New(403, "演示管理员受保护，不允许修改")
		}
		if admin && statusChange {
			super, err := hasSuperRole(ctx, tx, id)
			if err != nil {
				return err
			}
			if super {
				return errorx.New(10001, "不能通过接口修改超级管理员状态")
			}
		}
		// 固定字段白名单，防止角色、密码、账号状态等被资料接口批量赋值。
		columns := []string{"nickname", "avatar_url", "phone", "email", "age", "gender"}
		if statusChange {
			columns = []string{"status"}
		}
		data := map[string]any{}
		for _, column := range columns {
			if value, ok := fields[column]; ok {
				data[column] = value
			}
		}
		if len(data) == 0 {
			return errorx.New(10001, "请提供需要修改的字段")
		}
		data["updated_at"] = time.Now()
		_, err := accountModel(ctx, admin).WithTX(tx).WherePri(id).Data(data).Update()
		return err
	})
}

func accountProfileColumns(admin bool) string {
	if admin {
		return profileColumns + ",is_protected"
	}
	return profileColumns
}
func UpdateAdminPassword(ctx context.Context, id int, hash string) error {
	return db.Transaction(ctx, func(tx *db.TX) error {
		var row struct {
			ID          int
			IsProtected bool
		}
		if err := modeldao.SystemAdmin.Ctx(ctx).WithTX(tx).WherePri(id).WhereNull("deleted_at").LockUpdate().One(&row); err != nil {
			return err
		}
		if row.ID == 0 {
			return errorx.New(10001, "管理员不存在")
		}
		if row.IsProtected {
			return errorx.New(403, "演示管理员受保护，不允许修改密码")
		}
		now := time.Now()
		_, err := modeldao.SystemAdmin.Ctx(ctx).WithTX(tx).WherePri(id).Data(map[string]any{"password_hash": hash, "password_changed_at": now, "updated_at": now}).Update()
		return err
	})
}
