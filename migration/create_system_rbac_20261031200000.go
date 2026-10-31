package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&CreateSystemRBAC20261031200000{}) }

type CreateSystemRBAC20261031200000 struct{}

func (*CreateSystemRBAC20261031200000) Up(db *gorm.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS system_role (id int NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID', code varchar(64) NOT NULL COMMENT '唯一业务编码', name varchar(100) NOT NULL COMMENT '名称', status tinyint NOT NULL DEFAULT 1 COMMENT '状态：0停用，1启用', created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间', UNIQUE KEY uk_system_role_code(code)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='后台角色'`,
		`CREATE TABLE IF NOT EXISTS system_permission (id int NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID', code varchar(100) NOT NULL COMMENT '唯一业务编码', name varchar(100) NOT NULL COMMENT '名称', UNIQUE KEY uk_system_permission_code(code)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='后台接口权限点'`,
		`CREATE TABLE IF NOT EXISTS system_admin_role (admin_id int NOT NULL COMMENT '后台管理员ID，关联system_admin.id', role_id int NOT NULL COMMENT '角色ID，关联system_role.id', PRIMARY KEY(admin_id,role_id), KEY idx_system_admin_role_role(role_id), CONSTRAINT fk_system_ar_admin FOREIGN KEY(admin_id) REFERENCES system_admin(id), CONSTRAINT fk_system_ar_role FOREIGN KEY(role_id) REFERENCES system_role(id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员角色关联'`,
		`CREATE TABLE IF NOT EXISTS system_role_permission (role_id int NOT NULL COMMENT '角色ID，关联system_role.id', permission_id int NOT NULL COMMENT '接口权限ID，关联system_permission.id', PRIMARY KEY(role_id,permission_id), KEY idx_system_rp_permission(permission_id), CONSTRAINT fk_system_rp_role FOREIGN KEY(role_id) REFERENCES system_role(id), CONSTRAINT fk_system_rp_permission FOREIGN KEY(permission_id) REFERENCES system_permission(id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联'`,
		`INSERT INTO system_role(code,name,status) VALUES('super_admin','超级管理员',1) ON DUPLICATE KEY UPDATE code=VALUES(code)`,
		`INSERT INTO system_permission(code,name) VALUES('system.admin.read','查看管理员'),('system.role.read','查看角色'),('system.permission.read','查看权限点') ON DUPLICATE KEY UPDATE code=VALUES(code)`,
	}
	for _, sql := range statements {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}
