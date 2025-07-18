上次我们在[grpc基础]做了一个不用注册中心的grpc接口样例，本文将引入微服务的注册中心书写一套微服务grpc接口。

<h2 id="sGsHi">一、准备consul</h2>
<h3 id="LypMa">1.1. 下载consul</h3>
在mac上下载consul

> brew install consul
>

<h3 id="TYwSH">1.2. 启动consul</h3>
> consul agent -dev
>

<h3 id="V1zlE">1.3. 访问consu控制台</h3>
> [http://localhost:8500](http://localhost:8500)
>

<h2 id="zU6xG">二、代码实现</h2>
<h3 id="sGNy7">2.1 项目结构</h3>
```go
grpc-consul-demo/
├── client/
│   └── main.go          # 客户端入口
├── proto/
│    ├── hello.proto  # Proto定义
│    └── hello.pb.go  # 生成的代码
|    └── hello_grpc.pb.go  # 生成的代码
├── server/
│   ├── main.go          # 服务端入口
│   └── service/        # 业务实现
│       └── hello.go
├── pkg/
│   ├── discovery/       # 服务发现封装
│   │   └── consul.go
│   └── registration/    # 服务注册封装
│       └── consul.go
└── go.mod
```

<h3 id="CNWBQ">2.2 编写代码</h3>
<h4 id="zxLod">2.2.1. proto定义</h4>
```go
// 这里说明使用的是proto3语法
syntax = "proto3";

// 这部分的内容是关于最后生成的go文件是放在哪个目录哪个包中，.代表在当前目录生成，service代表了生成的go文件的包名是service.
option go_package = ".;service";

// 然后我们需要定义一个服务，在这个服务中需要有一个方法，这个方法可以接受客户端的参数，再返回服务响应。
//其实很容易可以看出，我们定义了一个service，称为SayHello,这个服务有一个rpc方法，名为SayHello
//这个方法会发送一个HelloRequest,然后返回一个HelloResponse
service SayHello{
  rpc SayHello(HelloRequest) returns(HelloResponse) {}
}

// message 关键字，可以理解为golang中的结构体
//这里比较体别的是变量后面的“赋值”。注意，这里并不是赋值，而是在这个定义这个变量在这个message中的位置
message HelloRequest {
  string requestName = 1;
}

message HelloResponse{
  string responseMsg = 1;
}
```

生成.pb.go 与hello_grpc.pb.go文件，命令如下：

> protoc --go_out=. --go-grpc_out=. proto/hello.proto
>



<h4 id="JDDJr">2.2.2 服务端实现</h4>
<h5 id="lgY8F">服务端入口(server/main.go)</h5>
```go
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

```

<h5 id="hTy1w">业务实现(server/service/hello.go)</h5>
```go
package service

import (
	"context"
	"fmt"

	hello "github.com/chaser_ZH/grpc-study/proto"
)

type HelloServer struct {
	hello.UnimplementedSayHelloServer
}

func (s *HelloServer) SayHello(ctx context.Context, req *hello.HelloRequest) (*hello.HelloResponse, error) {
	return &hello.HelloResponse{
		ResponseMsg: fmt.Sprintf("Hello, %s!", req.RequestName),
	}, nil
}

```

<h5 id="CJ7I0">consul注册封装（pkg/registeration/consul.go）</h5>
```go
package registeration

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"time"
)

type ConsulRegister struct {
	client    *api.Client
	serviceID string
}

func NewConsulRegister(consulAddr string) (*ConsulRegister, error) {
	config := api.DefaultConfig()
	config.Address = consulAddr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ConsulRegister{client: client}, nil
}

func (r *ConsulRegister) Register(serviceName string, port int) error {
	r.serviceID = fmt.Sprintf("%s-%d", serviceName, time.Now().Unix())

	reg := &api.AgentServiceRegistration{
		ID:      r.serviceID,
		Name:    serviceName,
		Port:    port,
		Address: getLocalIP(),
		Check: &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d", getLocalIP(), port),
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	return r.client.Agent().ServiceRegister(reg)
}

func (r *ConsulRegister) Deregister() error {
	if r.serviceID == "" {
		return nil
	}
	return r.client.Agent().ServiceDeregister(r.serviceID)
}

func getLocalIP() string {
	// 简化实现，生产环境需完善
	return "127.0.0.1"
}
```

<h4 id="eEoKQ">2.2.4 客户端实现</h4>
<h5 id="N6VQ1">客户端入口（client/main.go）</h5>
```go
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

```

<h5 id="G1wX5">服务发现与封装（pkg/discovery/consul.go）</h5>
```go
package discovery

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"math/rand"
)

type ConsulDiscoverer struct {
	client *api.Client
}

func NewConsulDiscoverer(consulAddr string) (*ConsulDiscoverer, error) {
	client, err := api.NewClient(&api.Config{
		Address: consulAddr,
	})
	if err != nil {
		return nil, err
	}
	return &ConsulDiscoverer{client: client}, nil
}

func (d *ConsulDiscoverer) Discover(serviceName string) (string, error) {
	entries, _, err := d.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("consul query failed: %v", err)
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("no healthy instances available")
	}

	// Go 1.20+ 推荐方式（无需显式Seed）
	selected := entries[rand.Intn(len(entries))] // 自动使用随机源
	return fmt.Sprintf("%s:%d", selected.Service.Address, selected.Service.Port), nil
}

```



<h2 id="th1rS">三 测试</h2>
<h3 id="PEYgs">3.1 启动consul</h3>
> consul agent -dev
>

<h3 id="dhFgT">3.2 启动server</h3>
```plain
cd server
go run main.go
```

<h3 id="MmBUX">3.3 启动client</h3>
> cd client
>
> go run main.go
>

