package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	_ "github.com/scienceol/studio/service/docs" // 导入自动生成的 docs 包
	"github.com/scienceol/studio/service/internal/configs/webapp"
	"github.com/scienceol/studio/service/pkg/core/notify/events"
	"github.com/scienceol/studio/service/pkg/middleware/db"
	"github.com/scienceol/studio/service/pkg/middleware/logger"
	"github.com/scienceol/studio/service/pkg/middleware/redis"
	"github.com/scienceol/studio/service/pkg/middleware/trace"
	"github.com/scienceol/studio/service/pkg/model/migrate"
	"github.com/scienceol/studio/service/pkg/utils"
	"github.com/scienceol/studio/service/pkg/web"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewWeb() *cobra.Command {
	webServer := &cobra.Command{
		Use:  "apiserver",
		Long: `api server start`,

		// stop printing usage when the command errors
		SilenceUsage:       true,
		PersistentPreRunE:  initGlobalResource,
		PersistentPostRunE: cleanGlobalResource,
		PreRunE:            initWeb,
		RunE:               newRouter,
		PostRunE:           cleanWebResource,
	}

	return webServer
}

func initGlobalResource(_ *cobra.Command, _ []string) error {
	// 初始化全局环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found - using environment variables")
	}

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

	return nil
}

func initMigrate(cmd *cobra.Command, _ []string) error {
	config := webapp.Config()
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

	return nil
}

func initWeb(cmd *cobra.Command, _ []string) error {
	config := webapp.Config()

	// Automatically generate Swagger documentation upon startup.
	logger.Infof(cmd.Context(), "📖 Attempting to generate Swagger documentation...")
	// The swag command needs to be run from the project root to correctly find the main.go file.
	swagCmd := exec.Command("swag", "init", "-g", "main.go")
	swagCmd.Dir = "." // Run from the project root
	output, err := swagCmd.CombinedOutput()
	if err != nil {
		// Log as a warning instead of a fatal error, so the server can still start
		// even if swag is not installed or fails.
		logger.Warnf(cmd.Context(), "❌ Could not generate Swagger documentation: %v. Output: %s", err, string(output))
	} else {
		logger.Infof(cmd.Context(), "✅ Swagger documentation generated successfully.")
	}

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

	// 检查数据库迁移状态
	if err := migrate.HandleAutoMigration(cmd.Context(), config); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	// 初始化 redis
	redis.InitRedis(cmd.Context(), &redis.Redis{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	return nil
}

func newRouter(cmd *cobra.Command, _ []string) error {
	router := gin.Default()

	web.NewRouter(cmd.Root().Context(), router)
	port := webapp.Config().Server.Port
	addr := ":" + strconv.Itoa(port)

	httpServer := http.Server{
		Addr:              ":" + strconv.Itoa(webapp.Config().Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	// 添加启动成功的日志输出
	fmt.Printf("🚀 Server starting on http://localhost:%d\n", port)
	fmt.Printf("📡 API Server is running at: http://0.0.0.0:%d\n", port)
	fmt.Printf("🔧 Server configuration: %+v\n", addr)

	// 异步监听端口
	utils.SafelyGo(func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				logger.Errorf(cmd.Context(), "start server err: %v\n", err)
			}
		}
	}, func(err error) {
		logger.Errorf(cmd.Context(), "run http server err: %+v", err)
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

func cleanWebResource(cmd *cobra.Command, _ []string) error {
	// FIXME: 关系消息通知中心
	// FIXME: 关闭 websocket
	events.NewEvents().Close(cmd.Context())
	redis.CloseRedis(cmd.Context())
	db.ClosePostgres(cmd.Context())
	trace.CloseTrace()
	return nil
}

func cleanGlobalResource(_ *cobra.Command, _ []string) error {
	// 服务退出清理资源
	logger.Close()
	return nil
}
