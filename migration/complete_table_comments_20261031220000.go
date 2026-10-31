package migration

import (
	"fmt"
	"go-gin/internal/migration"
	"gorm.io/gorm"
	"regexp"
	"strings"
)

func init() { migration.Register(&CompleteTableComments20261031220000{}) }

type CompleteTableComments20261031220000 struct{}

func (*CompleteTableComments20261031220000) Up(db *gorm.DB) error {
	return completeTableComments(db, "user", "system_admin", "system_role", "system_permission", "system_admin_role", "system_role_permission", "system_menu", "system_role_menu")
}

var tableDescriptions = map[string]string{
	"user": "用户端账号与资料", "system_admin": "后台管理员账号与资料", "system_role": "后台角色", "system_permission": "后台接口权限点",
	"system_admin_role": "后台管理员与角色关联", "system_role_permission": "后台角色与接口权限关联", "system_menu": "后台目录菜单按钮树", "system_role_menu": "后台角色与菜单节点关联",
}
var columnDescriptions = map[string]string{
	"id": "主键ID", "name": "名称", "code": "唯一业务编码", "status": "状态：0停用，1启用",
	"created_at": "创建时间", "updated_at": "更新时间", "deleted_at": "软删除时间，NULL表示未删除",
	"admin_id": "后台管理员ID，关联system_admin.id", "role_id": "角色ID，关联system_role.id",
	"permission_id": "接口权限ID，关联system_permission.id", "menu_id": "菜单节点ID，关联system_menu.id",
	"parent_id": "父节点ID，0表示根节点", "type": "节点类型：directory目录，menu菜单，button按钮",
	"path": "前端路由路径，不是后端接口地址", "component": "前端页面组件标识", "icon": "前端图标标识",
	"permission_code": "关联接口权限编码，仅描述关联，不自动授权", "sort": "排序值，越小越靠前，同值按ID排序",
	"age": "年龄，0表示未填写", "password_hash": "bcrypt密码哈希，旧账号未设置时为NULL",
	"nickname": "昵称", "avatar_url": "头像地址", "phone": "手机号资料，不代表已验证", "email": "邮箱资料，不代表已验证",
	"gender": "性别：0未知，1男，2女", "register_ip": "注册时直连客户端IP", "last_login_at": "最近成功登录时间，未登录为NULL",
	"last_login_ip": "最近成功登录的直连客户端IP", "login_count": "成功登录次数", "password_changed_at": "最近设置密码时间，未知为NULL",
}

// MySQL修改列注释需要完整列定义；从SHOW CREATE TABLE保留原定义，避免改动类型、默认值及自增属性。
func completeTableComments(db *gorm.DB, tables ...string) error {
	commentPattern := regexp.MustCompile(` COMMENT '(?:[^'\\]|\\.|'')*'`)
	for _, table := range tables {
		description, ok := tableDescriptions[table]
		if !ok {
			return fmt.Errorf("未定义表注释: %s", table)
		}
		rows, err := db.Raw("SHOW CREATE TABLE `" + table + "`").Rows()
		if err != nil {
			return err
		}
		var tableName, ddl string
		if !rows.Next() {
			rows.Close()
			return fmt.Errorf("未找到表 %s", table)
		}
		err = rows.Scan(&tableName, &ddl)
		rows.Close()
		if err != nil {
			return err
		}
		clauses := []string{"COMMENT='" + description + "'"}
		for _, line := range strings.Split(ddl, "\n") {
			definition := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), ","))
			if !strings.HasPrefix(definition, "`") {
				continue
			}
			end := strings.Index(definition[1:], "`")
			if end < 0 {
				continue
			}
			column := definition[1 : end+1]
			comment, ok := columnDescriptions[column]
			if !ok {
				return fmt.Errorf("%s.%s缺少字段注释定义", table, column)
			}
			if column == "name" && (table == "user" || table == "system_admin") {
				comment = "登录用户名，全表唯一"
			}
			if column == "code" && table == "system_permission" {
				comment = "接口权限标识，如system.admin.read"
			}
			definition = commentPattern.ReplaceAllString(definition, "")
			clauses = append(clauses, "MODIFY COLUMN "+definition+" COMMENT '"+comment+"'")
		}
		if err := db.Exec("ALTER TABLE `" + table + "` " + strings.Join(clauses, ", ")).Error; err != nil {
			return err
		}
	}
	return nil
}
