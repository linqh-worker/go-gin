package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"go-gin/internal/component/db"
	"go-gin/internal/errorx"
	modeldao "go-gin/model/dao"
	"go-gin/model/do"
	"time"
)

type ConfigItem struct {
	ID                                                                                           int64
	RecordType, GroupCode, GroupName, ConfigKey, Name, ConfigValue, ValueType, InputType, Remark string
	IsPublic, IsBuiltin                                                                          bool
	Sort                                                                                         int
	Version                                                                                      int64
	UpdatedAt                                                                                    *time.Time
}

func ConfigItems(ctx context.Context) ([]ConfigItem, error) {
	rows := []ConfigItem{}
	err := modeldao.SystemConfig.Ctx(ctx).Order("group_code,sort,id").All(&rows)
	return rows, err
}

type ConfigChange struct {
	Key, Value string
	Version    int64
}

func SaveConfigGroup(ctx context.Context, group string, actor int, changes []ConfigChange, validate func(ConfigItem, string) error) error {
	return configTransaction(ctx, func(tx *db.TX, groups map[string]ConfigItem) error {
		marker, exists := groups[group]
		if !exists {
			return errorx.New(10001, "配置分组不存在")
		}
		rows := []ConfigItem{}
		query := modeldao.SystemConfig.Ctx(ctx).WithTX(tx)
		if err := query.Where(do.SystemConfig{GroupCode: group, RecordType: "item"}).Order("id").LockUpdate().All(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			return errorx.New(10001, "配置分组不存在")
		}
		byKey := map[string]ConfigItem{}
		for _, r := range rows {
			byKey[r.ConfigKey] = r
		}
		// 先校验整批，再写入；任意错误都会回滚本组修改。
		for _, c := range changes {
			r, ok := byKey[c.Key]
			if !ok {
				return errorx.New(10001, "配置项不属于当前分组")
			}
			if r.Version != c.Version {
				return errorx.New(409, "配置已被其他人修改，请刷新后重试")
			}
			if err := validate(r, c.Value); err != nil {
				return err
			}
		}
		now := time.Now()
		changed := false
		for _, c := range changes {
			r := byKey[c.Key]
			if r.ConfigValue == c.Value {
				continue
			}
			n, err := query.Where(do.SystemConfig{Id: r.ID, Version: c.Version}).Data(do.SystemConfig{ConfigValue: c.Value, Version: c.Version + 1, UpdatedBy: actor, UpdatedAt: now}).Update()
			if err != nil {
				return err
			}
			changed = true
			if n != 1 {
				return errorx.New(409, "配置保存冲突，请刷新后重试")
			}
		}
		if changed {
			return bumpConfigGroup(ctx, tx, marker, actor)
		}
		return nil
	})
}

// 分组行先按ID统一加锁，再锁配置项；低频配置写入采用固定锁顺序。
func configTransaction(ctx context.Context, fn func(*db.TX, map[string]ConfigItem) error) error {
	err := db.Transaction(ctx, func(tx *db.TX) error {
		rows := []ConfigItem{}
		if err := modeldao.SystemConfig.Ctx(ctx).WithTX(tx).Where(do.SystemConfig{RecordType: "group"}).Order("id").LockUpdate().All(&rows); err != nil {
			return err
		}
		groups := map[string]ConfigItem{}
		for _, r := range rows {
			groups[r.GroupCode] = r
		}
		return fn(tx, groups)
	})
	var duplicate *mysql.MySQLError
	if errors.As(err, &duplicate) && duplicate.Number == 1062 {
		return errorx.New(10001, "分组编码或配置key已存在")
	}
	return err
}
func bumpConfigGroup(ctx context.Context, tx *db.TX, g ConfigItem, actor int) error {
	_, err := modeldao.SystemConfig.Ctx(ctx).WithTX(tx).WherePri(g.ID).Data(do.SystemConfig{Version: g.Version + 1, UpdatedBy: actor, UpdatedAt: time.Now()}).Update()
	return err
}
func CreateConfigGroup(ctx context.Context, code, name string, actor int) error {
	return configTransaction(ctx, func(tx *db.TX, groups map[string]ConfigItem) error {
		if _, ok := groups[code]; ok {
			return errorx.New(10001, "分组编码已存在")
		}
		now := time.Now()
		_, err := modeldao.SystemConfig.Ctx(ctx).WithTX(tx).Data(do.SystemConfig{RecordType: "group", GroupCode: code, GroupName: name, ConfigKey: "__group__." + code, Name: name, ConfigValue: "", CreatedBy: actor, UpdatedBy: actor, CreatedAt: now, UpdatedAt: now}).Insert()
		return err
	})
}
func UpdateConfigGroup(ctx context.Context, code, name string, version int64, actor int) error {
	return configTransaction(ctx, func(tx *db.TX, groups map[string]ConfigItem) error {
		g, ok := groups[code]
		if !ok {
			return errorx.New(10001, "配置分组不存在")
		}
		if g.Version != version {
			return errorx.New(409, "分组已变化，请刷新后重试")
		}
		var rows []ConfigItem
		q := modeldao.SystemConfig.Ctx(ctx).WithTX(tx)
		if err := q.Where(do.SystemConfig{GroupCode: code}).Order("id").LockUpdate().All(&rows); err != nil {
			return err
		}
		for _, r := range rows {
			data := do.SystemConfig{GroupName: name, Version: r.Version + 1, UpdatedBy: actor, UpdatedAt: time.Now()}
			if r.RecordType == "group" {
				data.Name = name
			}
			if _, err := q.WherePri(r.ID).Data(data).Update(); err != nil {
				return err
			}
		}
		return nil
	})
}
func DeleteConfigGroup(ctx context.Context, code string, version int64) error {
	return configTransaction(ctx, func(tx *db.TX, groups map[string]ConfigItem) error {
		g, ok := groups[code]
		if !ok {
			return errorx.New(10001, "配置分组不存在")
		}
		if g.Version != version {
			return errorx.New(409, "分组已变化，请刷新后重试")
		}
		var rows []ConfigItem
		q := modeldao.SystemConfig.Ctx(ctx).WithTX(tx)
		if err := q.Where(do.SystemConfig{GroupCode: code}).Order("id").LockUpdate().All(&rows); err != nil {
			return err
		}
		for _, r := range rows {
			if r.IsBuiltin {
				return errorx.New(10001, "分组包含内置关键配置，不允许删除")
			}
		}
		_, err := q.Where(do.SystemConfig{GroupCode: code}).Delete()
		return err
	})
}
func SaveConfigDefinition(ctx context.Context, id int64, item ConfigItem, actor int, validate func(ConfigItem, string) error) error {
	return configTransaction(ctx, func(tx *db.TX, groups map[string]ConfigItem) error {
		target, ok := groups[item.GroupCode]
		if !ok {
			return errorx.New(10001, "配置分组不存在")
		}
		q := modeldao.SystemConfig.Ctx(ctx).WithTX(tx)
		var old ConfigItem
		if id > 0 {
			if err := q.WherePri(id).Where(do.SystemConfig{RecordType: "item"}).LockUpdate().One(&old); err != nil {
				return err
			}
			if old.ID == 0 {
				return errorx.New(10001, "配置项不存在")
			}
			if old.Version != item.Version {
				return errorx.New(409, "配置项已变化，请刷新后重试")
			}
			if old.IsBuiltin && (old.ConfigKey != item.ConfigKey || old.GroupCode != item.GroupCode || old.ValueType != item.ValueType || !item.IsBuiltin) {
				return errorx.New(10001, "内置关键配置不能改key、分组、数据类型或取消保护")
			}
		}
		if err := validate(item, item.ConfigValue); err != nil {
			return err
		}
		now := time.Now()
		data := do.SystemConfig{RecordType: "item", GroupCode: target.GroupCode, GroupName: target.GroupName, ConfigKey: item.ConfigKey, Name: item.Name, ConfigValue: item.ConfigValue, ValueType: item.ValueType, InputType: item.InputType, IsPublic: item.IsPublic, IsBuiltin: item.IsBuiltin, Sort: item.Sort, Remark: item.Remark, UpdatedBy: actor, UpdatedAt: now}
		if id == 0 {
			data.Version = 1
			data.CreatedBy = actor
			data.CreatedAt = now
			if _, err := q.Data(data).Insert(); err != nil {
				return err
			}
		} else {
			data.Version = old.Version + 1
			if _, err := q.WherePri(id).Data(data).Update(); err != nil {
				return err
			}
		}
		if err := bumpConfigGroup(ctx, tx, target, actor); err != nil {
			return err
		}
		if id > 0 && old.GroupCode != target.GroupCode {
			previous, exists := groups[old.GroupCode]
			if !exists {
				return errorx.New(10001, "原分组不存在")
			}
			return bumpConfigGroup(ctx, tx, previous, actor)
		}
		return nil
	})
}
func DeleteConfigDefinition(ctx context.Context, id, version int64, actor int) error {
	return configTransaction(ctx, func(tx *db.TX, groups map[string]ConfigItem) error {
		q := modeldao.SystemConfig.Ctx(ctx).WithTX(tx)
		var r ConfigItem
		if err := q.WherePri(id).Where(do.SystemConfig{RecordType: "item"}).LockUpdate().One(&r); err != nil {
			return err
		}
		if r.ID == 0 {
			return errorx.New(10001, "配置项不存在")
		}
		if r.Version != version {
			return errorx.New(409, "配置项已变化，请刷新后重试")
		}
		if r.IsBuiltin {
			return errorx.New(10001, "内置关键配置不能删除")
		}
		g, exists := groups[r.GroupCode]
		if !exists {
			return errorx.New(10001, "配置分组不存在")
		}
		if _, err := q.WherePri(id).Where(do.SystemConfig{RecordType: "item"}).Delete(); err != nil {
			return err
		}
		return bumpConfigGroup(ctx, tx, g, actor)
	})
}
