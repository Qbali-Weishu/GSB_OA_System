package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/company/oa-leave-system/internal/config"
	"github.com/company/oa-leave-system/internal/database"
	"github.com/company/oa-leave-system/internal/router"
	"github.com/company/oa-leave-system/internal/worker"
	"go.uber.org/zap"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		panic("加载配置失败: " + err.Error())
	}

	// 初始化日志
	logger, err := buildLogger(cfg.App.Env)
	if err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer logger.Sync()

	// 初始化数据库连接
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}
	defer db.Close()

	// 执行数据库迁移
	if err := database.Migrate(db, cfg.Database.MigrationsPath); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 初始化 Redis 连接
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("连接 Redis 失败", zap.Error(err))
	}
	defer rdb.Close()

	// 初始化路由和所有依赖
	r := router.New(cfg, db, rdb, logger)

	// 启动通知后台工作线程
	notifyWorker := worker.NewNotificationWorker(db, rdb, logger)
	go notifyWorker.Run()

	// 启动 HTTP 服务
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("服务启动", zap.String("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP 服务启动失败", zap.Error(err))
		}
	}()

	// 等待退出信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("收到退出信号，开始优雅关闭")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	notifyWorker.Stop()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP 服务关闭异常", zap.Error(err))
	}
	logger.Info("服务已安全退出")
}

func buildLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
