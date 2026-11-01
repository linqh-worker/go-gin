package db

import (
	"context"
	"gorm.io/gorm"
)

type sqlTemplateKey struct{}

// GORM Scan会临时替换Logger为Recorder，绕过ParamsFilter。
// 在各执行链末尾保留实际SQL模板，由DBLog使用，避免Scan日志插入绑定值。
func registerSQLLogContext(database *gorm.DB) error {
	capture := func(tx *gorm.DB) {
		if tx.Statement != nil && tx.Statement.SQL.Len() > 0 {
			ctx := tx.Statement.Context
			if ctx == nil {
				ctx = context.Background()
			}
			tx.Statement.Context = context.WithValue(ctx, sqlTemplateKey{}, tx.Statement.SQL.String())
		}
	}
	registrars := []func() error{
		func() error { return database.Callback().Create().After("*").Register("go-gin:sql-template", capture) },
		func() error { return database.Callback().Query().After("*").Register("go-gin:sql-template", capture) },
		func() error { return database.Callback().Update().After("*").Register("go-gin:sql-template", capture) },
		func() error { return database.Callback().Delete().After("*").Register("go-gin:sql-template", capture) },
		func() error { return database.Callback().Row().After("*").Register("go-gin:sql-template", capture) },
		func() error { return database.Callback().Raw().After("*").Register("go-gin:sql-template", capture) },
	}
	for _, register := range registrars {
		if err := register(); err != nil {
			return err
		}
	}
	return nil
}
func sqlTrace(ctx context.Context, fc func() (string, int64)) (string, int64) {
	sql, rows := fc()
	if ctx != nil {
		if template, ok := ctx.Value(sqlTemplateKey{}).(string); ok && template != "" {
			sql = template
		}
	}
	return sql, rows
}
