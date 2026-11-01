package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&GroupSystemConfig20261101181000{}) }

type GroupSystemConfig20261101181000 struct{}

func (*GroupSystemConfig20261101181000) Up(database *gorm.DB) error {
	if !database.Migrator().HasColumn("system_config", "group_name") {
		if err := database.Exec("ALTER TABLE system_config ADD COLUMN group_name varchar(100) NOT NULL DEFAULT '' COMMENT '分组中文名称，同group_code保持一致，用于TAB显示' AFTER group_code").Error; err != nil {
			return err
		}
	}
	return database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("UPDATE system_config SET group_code='website',group_name='网站设置',sort=CASE config_key WHEN 'website.name' THEN 10 WHEN 'website.home_title' THEN 20 WHEN 'website.keywords' THEN 30 ELSE 40 END,version=version+1 WHERE config_key IN('website.name','website.home_title','website.keywords','website.description') AND (group_code<>'website' OR group_name<>'网站设置')").Error; err != nil {
			return err
		}
		for code, name := range map[string]string{"config_settings": "系统配置", "config_read": "查看系统配置", "config_update": "保存系统配置"} {
			if err := tx.Table("system_menu").Where("code = ?", code).UpdateColumn("name", name).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
