package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&CreateSystemConfig20261101180000{}) }

type CreateSystemConfig20261101180000 struct{}

func (*CreateSystemConfig20261101180000) Up(database *gorm.DB) error {
	if err := database.Exec("CREATE TABLE IF NOT EXISTS system_config (id bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',group_code varchar(64) NOT NULL COMMENT '分组编码，对应设置TAB',config_key varchar(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL COMMENT '全局唯一配置键，发布后保持稳定',name varchar(100) NOT NULL COMMENT '配置中文名称',config_value text NOT NULL COMMENT '配置值，按value_type校验，空字符串为有效值',value_type varchar(20) NOT NULL DEFAULT 'string' COMMENT '数据类型：string integer boolean json',input_type varchar(20) NOT NULL DEFAULT 'input' COMMENT '控件：input textarea number switch json',dict_code varchar(64) DEFAULT NULL COMMENT '预留关联字典编码，当前不使用',is_public tinyint unsigned NOT NULL DEFAULT 0 COMMENT '是否允许匿名公开：0否1是',is_builtin tinyint unsigned NOT NULL DEFAULT 0 COMMENT '内置配置：禁止删除和修改键，允许改值',sort int NOT NULL DEFAULT 0 COMMENT '组内显示顺序',remark varchar(500) NOT NULL DEFAULT '' COMMENT '说明',version bigint NOT NULL DEFAULT 1 COMMENT '乐观锁版本，保存时须匹配',created_by bigint DEFAULT NULL COMMENT '创建管理员ID，迁移初始化为NULL',updated_by bigint DEFAULT NULL COMMENT '最近修改管理员ID',created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',PRIMARY KEY(id),UNIQUE KEY uk_system_config_key(config_key),KEY idx_system_config_group_sort(group_code,sort,id)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统动态配置，敏感连接凭据不在此表维护'").Error; err != nil {
		return err
	}
	return database.Transaction(func(tx *gorm.DB) error {
		items := []struct {
			Group, Key, Name, Value, Input, Remark string
			Sort                                   int
		}{
			{"website", "website.name", "网站名称", "Go Gin 管理后台", "input", "用于登录页、侧栏与浏览器标题", 10},
			{"website", "website.home_title", "首页标题", "管理工作台", "input", "用于Dashboard工作台标题", 20},
			{"seo", "website.keywords", "关键字", "", "input", "多个关键字用英文逗号分隔", 10},
			{"seo", "website.description", "网站描述", "", "textarea", "用于网页description元信息", 20},
		}
		for _, v := range items {
			if err := tx.Exec("INSERT INTO system_config(group_code,config_key,name,config_value,value_type,input_type,is_public,is_builtin,sort,remark) SELECT ?,?,?,?,'string',?,1,1,?,? WHERE NOT EXISTS(SELECT 1 FROM (SELECT config_key FROM system_config) AS existing WHERE config_key=?)", v.Group, v.Key, v.Name, v.Value, v.Input, v.Sort, v.Remark, v.Key).Error; err != nil {
				return err
			}
		}
		var parent struct{ ID int }
		if err := tx.Table("system_menu").Where("code='system_root'").Take(&parent).Error; err != nil {
			return err
		}
		nodes := []struct {
			Code, Name, Type, Path, Component, Method, APIPath string
			Parent, Sort                                       int
		}{
			{"config_settings", "系统设置", "menu", "/system/settings", "system/config/index", "", "", parent.ID, 40},
		}
		for _, n := range nodes {
			if err := tx.Exec("INSERT INTO system_menu(parent_id,code,name,type,path,component,icon,sort,status) SELECT ?,?,?,?,?,?,'ri:settings-3-line',?,1 WHERE NOT EXISTS(SELECT 1 FROM (SELECT code FROM system_menu) AS existing WHERE code=?)", n.Parent, n.Code, n.Name, n.Type, n.Path, n.Component, n.Sort, n.Code).Error; err != nil {
				return err
			}
		}
		parent.ID = 0
		if err := tx.Table("system_menu").Where("code='config_settings'").Take(&parent).Error; err != nil {
			return err
		}
		for _, n := range []struct {
			Code, Name, Method, Path string
			Sort                     int
		}{{"config_read", "查看系统设置", "GET", "/system/configs", 10}, {"config_update", "保存系统设置", "PUT", "/system/configs/:group", 20}} {
			if err := tx.Exec("INSERT INTO system_menu(parent_id,code,name,type,method,api_path,sort,status) SELECT ?,?,?,'button',?,?,?,1 WHERE NOT EXISTS(SELECT 1 FROM (SELECT code FROM system_menu) AS existing WHERE code=?)", parent.ID, n.Code, n.Name, n.Method, n.Path, n.Sort, n.Code).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
