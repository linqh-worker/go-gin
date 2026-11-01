// system_menu 当前结构初始化，只建不存在的表，不修改已有结构或数据。
package migration

import (
	"go-gin/internal/migration"

	"gorm.io/gorm"
)

func init() { migration.Register(&CreateSystemMenu20261101130300{}) }

type CreateSystemMenu20261101130300 struct{}

func (*CreateSystemMenu20261101130300) Up(db *gorm.DB) error {
	return db.Exec(
		"CREATE TABLE IF NOT EXISTS `system_menu` (\n" +
			"  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID',\n" +
			"  `parent_id` int NOT NULL DEFAULT '0' COMMENT '父节点ID，0表示根节点',\n" +
			"  `code` varchar(64) NOT NULL COMMENT '唯一业务编码',\n" +
			"  `name` varchar(100) NOT NULL COMMENT '名称',\n" +
			"  `type` varchar(16) NOT NULL COMMENT '节点类型：directory目录，menu菜单，button按钮，api接口动作',\n" +
			"  `path` varchar(255) NOT NULL DEFAULT '' COMMENT '前端路由路径，不是后端接口地址',\n" +
			"  `component` varchar(255) NOT NULL DEFAULT '' COMMENT '前端页面组件标识',\n" +
			"  `method` varchar(10) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '' COMMENT '关联接口请求方法，空值表示纯展示节点',\n" +
			"  `api_path` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '关联接口路由模板，如/system/users/:id；path仍为前端页面路径',\n" +
			"  `icon` varchar(100) NOT NULL DEFAULT '' COMMENT '前端图标标识',\n" +
			"  `sort` int NOT NULL DEFAULT '0' COMMENT '排序值，越小越靠前，同值按ID排序',\n" +
			"  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：0停用，1启用',\n" +
			"  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n" +
			"  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n" +
			"  PRIMARY KEY (`id`),\n" +
			"  UNIQUE KEY `uk_system_menu_code` (`code`),\n" +
			"  KEY `idx_system_menu_parent` (`parent_id`,`sort`)\n" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='后台目录菜单按钮树'",
	).Error
}
