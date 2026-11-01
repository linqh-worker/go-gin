package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&ManageSystemConfig20261101190000{}) }

type ManageSystemConfig20261101190000 struct{}

func (*ManageSystemConfig20261101190000) Up(database *gorm.DB) error {
	if !database.Migrator().HasColumn("system_config", "record_type") {
		if err := database.Exec("ALTER TABLE system_config ADD COLUMN record_type varchar(16) NOT NULL DEFAULT 'item' COMMENT '记录类型：group分组，item配置项；分组占位以支持空组' AFTER id").Error; err != nil {
			return err
		}
	}
	return database.Transaction(func(tx *gorm.DB) error {
		var groups []struct{ GroupCode, GroupName string }
		if err := tx.Table("system_config").Select("group_code,MAX(group_name) AS group_name").Where("record_type='item'").Group("group_code").Scan(&groups).Error; err != nil {
			return err
		}
		for _, g := range groups {
			label := g.GroupName
			if label == "" {
				label = g.GroupCode
			}
			builtin := 0
			if g.GroupCode == "website" {
				builtin = 1
			}
			if err := tx.Exec("INSERT INTO system_config(record_type,group_code,group_name,config_key,name,config_value,is_builtin) SELECT 'group',?,?,?,?, '',? WHERE NOT EXISTS(SELECT 1 FROM (SELECT config_key FROM system_config) AS existing WHERE config_key=?)", g.GroupCode, label, "__group__."+g.GroupCode, label, builtin, "__group__."+g.GroupCode).Error; err != nil {
				return err
			}
		}
		var parent struct{ ID int }
		if err := tx.Table("system_menu").Where("code='config_settings'").Take(&parent).Error; err != nil {
			return err
		}
		for i, n := range []struct{ Code, Name, Method, Path string }{
			{"config_group_create", "新增配置分组", "POST", "/system/config-groups"}, {"config_group_update", "编辑配置分组", "PUT", "/system/config-groups/:group"}, {"config_group_delete", "删除配置分组", "DELETE", "/system/config-groups/:group"},
			{"config_item_create", "新增配置项", "POST", "/system/config-items"}, {"config_item_update", "编辑配置项", "PUT", "/system/config-items/:id"}, {"config_item_delete", "删除配置项", "DELETE", "/system/config-items/:id"},
		} {
			if err := tx.Exec("INSERT INTO system_menu(parent_id,code,name,type,method,api_path,sort,status) SELECT ?,?,?,'button',?,?,?,1 WHERE NOT EXISTS(SELECT 1 FROM (SELECT code FROM system_menu) AS existing WHERE code=?)", parent.ID, n.Code, n.Name, n.Method, n.Path, 30+i*10, n.Code).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
