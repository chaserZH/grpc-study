package main

import (
	"context"
	"github.com/chaser_ZH/grpc-study/pkg/discovery"
	hello "github.com/chaser_ZH/grpc-study/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	// 无注册中心启动
	//NoRegisterCenterStart()
	// 注册中心启动
	RegisterCenterStart()
}

func NoRegisterCenterStart() {
	// 1. 创建gRpc客户端连接(推荐全局复用)
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials((insecure.NewCredentials())))
	if err != nil {
		log.Fatal("failed to create grpc client: %v", err)
	}
	defer conn.Close()

	// 2. 创建客户端
	client := hello.NewSayHelloClient(conn)

	// 3. 调用服务端方法（使用context设置超时）
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.SayHello(ctx, &hello.HelloRequest{RequestName: "World"})
	if err != nil {
		log.Fatalf("RPC failed: %v", err)
	}
	log.Printf("服务端响应: %v", resp.ResponseMsg)
}

func RegisterCenterStart() {
	// 1. 从Consul发现服务
	consulAddr := "127.0.0.1:8500"
	serviceName := "hello-service"

	discoverer, err := discovery.NewConsulDiscoverer(consulAddr)
	if err != nil {
		log.Fatalf("failed to create discoverer: %v", err)
	}

	addr, err := discoverer.Discover(serviceName)
	if err != nil {
		log.Fatalf("discovery failed: %v", err)
	}
	log.Printf("discovered service address: %s", addr)

	// 2. 建立gRPC连接
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
	)
	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	defer conn.Close()

	// 3. 创建客户端
	client := hello.NewSayHelloClient(conn)

	// 4. 调用服务
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &hello.HelloRequest{RequestName: "World"})
	if err != nil {
		log.Fatalf("RPC failed: %v", err)
	}

	log.Printf("Response: %s", resp.ResponseMsg)
}
