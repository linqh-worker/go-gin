package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&CreateSystemMenu20261031210000{}) }

type CreateSystemMenu20261031210000 struct{}

func (*CreateSystemMenu20261031210000) Up(db *gorm.DB) error {
	for _, sql := range []string{
		`CREATE TABLE IF NOT EXISTS system_menu (id int NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID', parent_id int NOT NULL DEFAULT 0 COMMENT '父节点ID，0表示根节点', code varchar(64) NOT NULL COMMENT '唯一业务编码', name varchar(100) NOT NULL COMMENT '名称', type varchar(16) NOT NULL COMMENT '节点类型：directory目录，menu菜单，button按钮', path varchar(255) NOT NULL DEFAULT '' COMMENT '前端路由路径，不是后端接口地址', component varchar(255) NOT NULL DEFAULT '' COMMENT '前端页面组件标识', icon varchar(100) NOT NULL DEFAULT '' COMMENT '前端图标标识', permission_code varchar(100) NOT NULL DEFAULT '' COMMENT '关联接口权限编码，仅描述关联，不自动授权', sort int NOT NULL DEFAULT 0 COMMENT '排序值，越小越靠前，同值按ID排序', status tinyint NOT NULL DEFAULT 1 COMMENT '状态：0停用，1启用', created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间', UNIQUE KEY uk_system_menu_code(code), KEY idx_system_menu_parent(parent_id,sort)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='后台目录菜单按钮树，parent_id=0为根节点'`,
		`CREATE TABLE IF NOT EXISTS system_role_menu (role_id int NOT NULL COMMENT '角色ID，关联system_role.id', menu_id int NOT NULL COMMENT '菜单节点ID，关联system_menu.id', PRIMARY KEY(role_id,menu_id), KEY idx_system_rm_menu(menu_id), CONSTRAINT fk_system_rm_role FOREIGN KEY(role_id) REFERENCES system_role(id), CONSTRAINT fk_system_rm_menu FOREIGN KEY(menu_id) REFERENCES system_menu(id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色菜单授权'`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}
