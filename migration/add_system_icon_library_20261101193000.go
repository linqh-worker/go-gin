package migration

import (
	"go-gin/internal/migration"
	"gorm.io/gorm"
)

func init() { migration.Register(&AddSystemIconLibrary20261101193000{}) }

type AddSystemIconLibrary20261101193000 struct{}

func (*AddSystemIconLibrary20261101193000) Up(database *gorm.DB) error {
	return database.Transaction(func(tx *gorm.DB) error {
		var parent struct{ ID int }
		if err := tx.Table("system_menu").Where("code='system_root'").Take(&parent).Error; err != nil {
			return err
		}
		return tx.Exec("INSERT INTO system_menu(parent_id,code,name,type,path,component,icon,sort,status) SELECT ?, 'icon_library', '图标库', 'menu', '/system/icons', 'system/icon/index', 'ri:palette-line', 50, 1 WHERE NOT EXISTS(SELECT 1 FROM (SELECT code FROM system_menu) AS existing WHERE code='icon_library')", parent.ID).Error
	})
}
