package router

import (
	"fmt"
	apiRouter "go-gin/app/api/router"
	systemRouter "go-gin/app/system/router"
	"go-gin/internal/httpx"
)

func Init(route *httpx.Engine) {
	_ = InitModule(route, "all")
}

// InitModule 只挂载选定端的路由；公共存活检查始终保留。
func InitModule(route *httpx.Engine, module string) error {
	if err := ValidateModule(module); err != nil {
		return err
	}
	RegisterCommonRoutes(route)
	if module == "all" || module == "api" {
		apiRouter.Register(route.Group("/api"))
	}
	if module == "all" || module == "system" {
		systemRouter.Register(route.Group("/system"))
	}
	return nil
}

func ValidateModule(module string) error {
	if module != "all" && module != "api" && module != "system" {
		return fmt.Errorf("module 只能为 all、api 或 system，当前为 %q", module)
	}
	return nil
}
