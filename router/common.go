package router

import (
	"go-gin/internal/component/db"
	"go-gin/internal/component/redisx"
	"go-gin/internal/errorx"
	"go-gin/internal/g"
	"go-gin/internal/httpx"
	modeldao "go-gin/model/dao"
	"go-gin/util"
)

func RegisterCommonRoutes(route *httpx.Engine) {
	route.NoMethod(func(ctx *httpx.Context) (any, error) {
		return nil, errorx.ErrMethodNotAllowed
	})
	route.NoRoute(func(ctx *httpx.Context) (any, error) {
		return nil, errorx.ErrNoRoute
	})
	// 两端均提供公开网站配置，私有配置不参与查询。
	route.GET("/public/config", func(ctx *httpx.Context) (any, error) {
		ctx.Header("Cache-Control", "no-store")
		values, err := modeldao.SystemConfig.PublicValues(ctx.Request.Context())
		if err != nil {
			return nil, errorx.NewServerError(503)
		}
		return values, nil
	})
	// 健康检测
	route.GET("/status", func(ctx *httpx.Context) (any, error) {
		db_err := db.WithContext(ctx).Ping()
		redis_err := redisx.Client().Ping(ctx).Err()
		return g.MapStrStr{
			"database": util.When(db_err == nil, "ok", "failed"),
			"redis":    util.When(redis_err == nil, "ok", "failed"),
		}, nil
	})
	route.GET("/", func(ctx *httpx.Context) (any, error) {
		return g.MapStrStr{"service": "go-gin", "api": "/api", "system": "/system"}, nil
	})
}
