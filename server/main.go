package main

import (
	"fmt"
	registration "github.com/chaser_ZH/grpc-study/pkg/registeration"
	hello "github.com/chaser_ZH/grpc-study/proto"
	"github.com/chaser_ZH/grpc-study/server/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	serviceName = "hello-service"
	servicePort = 50051
)

func main() {

	// 默认不注册中心
	//NoRegisterCenterStart()

	// 注册中心启动
	RegisterCenterStart()

}

func RegisterCenterStart() {
	// 1. 创建gRPC服务器
	srv := grpc.NewServer()

	// 2. 注册服务
	hello.RegisterSayHelloServer(srv, &service.HelloServer{})

	// 3. 添加健康检查
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// 4. 启动TCP监听
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 5. 注册到Consul
	consulReg, err := registration.NewConsulRegister("localhost:8500")
	if err != nil {
		log.Fatalf("failed to create consul register: %v", err)
	}
	defer consulReg.Deregister()

	err = consulReg.Register(serviceName, servicePort)
	if err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	// 6. 启动服务
	go func() {
		log.Printf("server started on port %d", servicePort)
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 7. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	srv.GracefulStop()
	log.Println("server stopped")
}

func NoRegisterCenterStart() {
	// 开启端口
	listen, _ := net.Listen("tcp", ":8080")

	//创建grpc服务
	grpcServer := grpc.NewServer()

	// 在grpc服务端中注册我们自己编写的服务
	hello.RegisterSayHelloServer(grpcServer, &service.HelloServer{})

	//启动服务
	err := grpcServer.Serve(listen)

	if err != nil {
		panic(err)
	}
}
