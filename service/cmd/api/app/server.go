package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/trace"
	"github.com/scienceol/studio/service/pkg/repository/db"
	"github.com/scienceol/studio/service/pkg/repository/redis"
	"github.com/scienceol/studio/service/pkg/utils"
	"github.com/scienceol/studio/service/pkg/web"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewWeb() *cobra.Command {

	rootCommand := &cobra.Command{
		Use:  "apiserver",
		Long: `api server start`,

		// stop printing usage when the command errors
		SilenceUsage:       true,
		PersistentPreRunE:  initGlobalResource,
		PreRunE:            nil,
		RunE:               newRouter,
		PostRunE:           nil,
		PersistentPostRunE: cleanGlobalResrource,
	}

	rootCommand.SetContext(utils.SetupSignalContext())

	return rootCommand
}

func initGlobalResource(cmd *cobra.Command, args []string) error {
	// 初始化全局环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found - using environment variables")
	}

	fmt.Println(os.LookupEnv("SERVER_PORT"))
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.AutomaticEnv()

	config := webapp.Config()
	if err := v.Unmarshal(config); err != nil {
		log.Fatal(err)
	}

	// 日志初始化
	logger.Init(&logger.LogConfig{
		Path:     config.Log.LogPath,
		LogLevel: config.Log.LogLevel,
		ServiceEnv: logger.ServiceEnv{
			Platform: config.Server.Platform,
			Service:  config.Server.Service,
			Env:      config.Server.Env,
		},
	})

	// 初始化 trace
	trace.InitTrace(cmd.Context(), &trace.TraceConfig{
		ServiceName:     fmt.Sprintf("%s-%s", config.Server.Service, config.Server.Platform),
		Version:         config.Trace.Version,
		TraceEndpoint:   config.Trace.TraceEndpoint,
		MetricEndpoint:  config.Trace.MetricEndpoint,
		TraceProject:    config.Trace.TraceProject,
		TraceInstanceID: config.Trace.TraceInstanceID,
		TraceAK:         config.Trace.TraceAK,
		TraceSK:         config.Trace.TraceSK,
	})

	// 初始化数据库
	db.InitPostgres(cmd.Context(), &db.Config{
		Host:   config.Database.Host,
		Port:   config.Database.Port,
		User:   config.Database.User,
		PW:     config.Database.Password,
		DBName: config.Database.Name,
		LogConf: db.LogConf{
			Level: config.Log.LogLevel,
		},
	})

	// 初始化 redis
	redis.InitRedis(cmd.Context(), &redis.Redis{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	return nil
}

func cleanGlobalResrource(cmd *cobra.Command, args []string) error {
	// 服务退出清理资源
	redis.CloseRedis(cmd.Context())
	db.ClosePostgres(cmd.Context())
	trace.CloseTrace()
	logger.Close()
	return nil
}

func newRouter(cmd *cobra.Command, args []string) error {
	router := gin.Default()

	web.NewRouter(router)

	port := webapp.Config().Server.Port
	addr := ":" + strconv.Itoa(port)

	httpServer := http.Server{
		Addr:              ":" + strconv.Itoa(webapp.Config().Server.Port),
		Handler:           router.Handler(),
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		TLSNextProto: func() map[string]func(*http.Server, *tls.Conn, http.Handler) {
			return make(map[string]func(*http.Server, *tls.Conn, http.Handler))
		}(),
	}

	// 添加启动成功的日志输出
	fmt.Printf("🚀 Server starting on http://localhost:%d\n", port)
	fmt.Printf("📡 API Server is running at: http://0.0.0.0:%d\n", port)
	fmt.Printf("🔧 Server configuration: %+v\n", addr)

	// 异步监听端口
	utils.SafelyGo(func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("start server err: %v\n", err)
			}
		}
	}, func(err error) {
		os.Exit(1)
	})

	// 服务启动成功提示
	fmt.Printf("✅ Server successfully started on port %d\n", port)
	fmt.Println("Press Ctrl+C to gracefully shutdown the server...")

	// 阻塞等待收到中断信号
	<-cmd.Context().Done()

	// 平滑超时退出
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("shut down server err: %+v", err)
	}
	return nil
}
