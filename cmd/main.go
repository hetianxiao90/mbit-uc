package main

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"uc/internal/protoc"
	"uc/internal/rpc"
	"uc/pkg/email"
	"uc/pkg/jwt"
	"uc/pkg/logger"
	"uc/pkg/mysql"
	"uc/pkg/nacos"
	"uc/pkg/rabbitmq"
	"uc/pkg/redis"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nacos.Init()
	nacos.InitConfig()
	logger.Init()
	email.Init()
	mysql.Init()
	redis.Init()
	jwt.Init()
	rabbitmq.Init()
	go rabbitmq.SendEmailStart()
	// 路由初始化
	//r := router.Init()
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", nacos.Config.App.Port))
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	protoc.RegisterUcServer(s, &rpc.UserRpc{})
	protoc.RegisterPublicServer(s, &rpc.PublicRpc{})
	nacos.RegisterInstance()
	fmt.Println("Serving start...")
	err = s.Serve(listen)
	if err != nil {
		nacos.DeregisterInstance()
		fmt.Printf("failed to serve: %v", err)
		return
	}
	//// 服务启动
	//srv := &http.Server{
	//	Addr:    ":" + viper.GetString("app.port"),
	//	Handler: r,
	//}
	//go func() {
	//	nacos.RegisterInstance()
	//	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	//		panic(err)
	//	}
	//
	//}()

	// -----------------------------优雅退出 -----------------------------
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞
	<-quit

	// 注销服务
	nacos.DeregisterInstance()

	// 关闭http
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	//if err := srv.Shutdown(ctx); err != nil {
	//	logger.Logger.Error("server shutdown err:", err)
	//	fmt.Printf("server shutdown: %v ", err)
	//}
	s.Stop()
	listen.Close()
	// 关闭DB
	if err = mysql.DBS.Close(); err != nil {
		logger.Logger.Error("DBS Close err:", err)
		fmt.Printf("DBS Close: %v", err)
	}

	// 关闭redis
	if err = redis.Client.Close(); err != nil {
		logger.Logger.Error("redis Close err:", err)
		fmt.Printf("redis Close: %v", err)
	}

	// 关闭rabbitmq
	rabbitmq.AMQP.Close()
	fmt.Println("Server exited gracefully")

}
