package router

import (
	apiRouter "go-gin/app/api/router"
	systemRouter "go-gin/app/system/router"
	"go-gin/internal/httpx"
)

func Init(route *httpx.Engine) {
	RegisterCommonRoutes(route)
	apiRouter.Register(route.Group("/api"))
	systemRouter.Register(route.Group("/system"))
}
