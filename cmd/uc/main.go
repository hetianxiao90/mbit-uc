package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"uc/configs"
	"uc/internal/logger"
	"uc/internal/mysql"
	"uc/internal/rabbitmq"
	"uc/internal/redis"
	"uc/internal/router"
	"uc/internal/util/email"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	// 配置初始化
	configs.Init()

	// 邮箱初始化
	email.Init()

	// 日志初始化
	logger.Init()

	// mysql初始化
	mysql.Init()

	// redis 初始化
	redis.Init()

	// rabbitmq初始化
	rabbitmq.Init()
	go rabbitmq.SendEmailStart()
	// 路由初始化
	r := router.Init()

	// 服务启动
	srv := &http.Server{
		Addr:    ":" + viper.GetString("app.port"),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("listen: %s", err)
		}
	}()

	// -----------------------------优雅退出 -----------------------------
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞
	<-quit

	// 关闭http
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown err:", err)
		fmt.Printf("server shutdown: %v ", err)
	}

	// 关闭DB
	if err := mysql.DBS.Close(); err != nil {
		logger.Error("DBS Close err:", err)
		fmt.Printf("DBS Close: %v", err)
	}

	// 关闭redis
	if err := redis.Client.Close(); err != nil {
		logger.Error("redis Close err:", err)
		fmt.Printf("redis Close: %v", err)
	}

	// 关闭rabbitmq
	rabbitmq.AMQP.Close()
	fmt.Println("Server exited gracefully")

}
