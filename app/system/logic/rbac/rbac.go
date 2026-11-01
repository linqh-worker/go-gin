package rbac

import (
	"context"
	"go-gin/app/system/dao"
	request "go-gin/app/system/request/rbac"
	"go-gin/internal/errorx"
	"regexp"
	"strings"
)

func Roles(ctx context.Context) ([]dao.Role, error)         { return dao.Roles(ctx) }
func AdminRoles(ctx context.Context, id int) ([]int, error) { return dao.AdminRoles(ctx, id) }
func SetAdminRoles(ctx context.Context, id int, roles []int) error {
	return dao.SetAdminRoles(ctx, id, roles)
}
func Admins(ctx context.Context) ([]dao.AdminSummary, error) { return dao.Admins(ctx) }

func SaveRole(ctx context.Context, id int, r request.RoleReq) error {
	r.Code = strings.TrimSpace(r.Code)
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" || r.Status == nil {
		return errorx.New(10001, "角色名称和状态不能为空")
	}
	if id == 0 && (!regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`).MatchString(r.Code) || r.Code == "super_admin") {
		return errorx.New(10001, "角色编码须为3至64位小写字母数字下划线，且不可使用保留编码")
	}
	return dao.SaveRole(ctx, id, r.Code, r.Name, *r.Status)
}
