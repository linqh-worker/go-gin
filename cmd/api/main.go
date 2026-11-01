package main

import (
	"flag"
	"fmt"
	"go-gin/config"
	"go-gin/event"
	"go-gin/internal/component/db"
	"go-gin/internal/component/logx"
	"go-gin/internal/component/redisx"
	"go-gin/internal/environment"
	"go-gin/internal/errorx"
	"go-gin/internal/httpx"
	"go-gin/internal/httpx/validators"
	"go-gin/internal/queue"
	_ "go-gin/internal/util"
	"go-gin/router"
	"go-gin/util"
	"net"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hibiken/asynqmon"
)

var configFile = flag.String("f", "./.env", "the config file")
var module = flag.String("module", "all", "加载模块：all、api、system")
var portOverride = flag.String("port", "", "覆盖监听地址，如 :8088；默认使用配置 app.port")

func main() {

	flag.Parse()
	if err := router.ValidateModule(*module); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	config.Init(*configFile)
	port := config.GetAppConf().Port
	if *portOverride != "" {
		port = *portOverride
	}
	_, portText, err := net.SplitHostPort(port)
	portNumber, numberErr := strconv.Atoi(portText)
	if err != nil || numberErr != nil || portNumber < 1 || portNumber > 65535 {
		fmt.Fprintln(os.Stderr, "监听地址无效，请使用 :8088 或 127.0.0.1:8088 这样的地址")
		os.Exit(1)
	}
	config.InitGlobalVars()
	config.InitEnvironment()

	logx.SetDefaultModule(*module)
	logx.InitConfig(config.GetLogConf())
	logx.Init()

	db.InitConfig(config.GetDbConf())
	db.Init()

	redisx.InitConfig(config.GetRedisConf())
	redisx.Init()

	queue.Init(config.GetRedisConf())
	defer queue.Close()

	event.Init()

	// 初始化第三方服务地址
	config.InitSvc()

	// 初始化http服务
	engine := initHttpServer()
	// 挂载监控
	if *module != "api" {
		mountMonitor(engine)
	}

	// 启动http服务
	fmt.Printf("Starting server: module=%s address=%s\n", *module, port)
	if err := engine.Run(port); err != nil {
		fmt.Fprintf(os.Stderr, "Start server error,err=%v\n", err)
		os.Exit(1)
	}
}

// 初始化http服务
func initHttpServer() *httpx.Engine {
	if environment.IsDebugMode() {
		httpx.SetDebugMode()
	} else {
		httpx.SetReleaseMode()
	}
	engine := httpx.Default()
	validators.Init()
	if err := router.InitModule(engine, *module); err != nil {
		panic(err)
	}
	return engine
}

// 挂载监控
func mountMonitor(engine *httpx.Engine) {
	// 挂载队列监控web ui
	mon := asynqmon.New(asynqmon.Options{
		RootPath:     "/monitor/queue",
		RedisConnOpt: queue.RedisClientOpt(config.GetRedisConf()),
	})

	r := engine.Engine.Group("/monitor")
	r.Use(func(ctx *gin.Context) {
		clientIp := ctx.ClientIP()
		// 如果客户端ip不在白名单内，直接返回403
		if !util.InArray(clientIp, config.GetMonitorConf().WhiteIpList) {
			httpx.Error(httpx.NewContext(ctx), errorx.ErrForbidden)
			ctx.Abort()
			return
		}
		ctx.Next()
	})
	r.GET("/queue/*any", gin.WrapH(mon))
}
