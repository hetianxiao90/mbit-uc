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
	"uc/internal/router"
	"uc/pkg/email"
	"uc/pkg/jwt"
	"uc/pkg/logger"
	"uc/pkg/mysql"
	"uc/pkg/rabbitmq"
	"uc/pkg/redis"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	configs.Init()
	email.Init()
	logger.Init()
	mysql.Init()
	redis.Init()
	jwt.Init()
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
			panic(err)
		}
	}()

	// -----------------------------优雅退出 -----------------------------
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞
	<-quit

	// 关闭http
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error("server shutdown err:", err)
		fmt.Printf("server shutdown: %v ", err)
	}

	// 关闭DB
	if err := mysql.DBS.Close(); err != nil {
		logger.Logger.Error("DBS Close err:", err)
		fmt.Printf("DBS Close: %v", err)
	}

	// 关闭redis
	if err := redis.Client.Close(); err != nil {
		logger.Logger.Error("redis Close err:", err)
		fmt.Printf("redis Close: %v", err)
	}

	// 关闭rabbitmq
	rabbitmq.AMQP.Close()

	fmt.Println("Server exited gracefully")

}
