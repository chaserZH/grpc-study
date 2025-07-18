<h1 id="pRSB7">一、安装proto</h1>


1. 安装proto

> brew install protobuf
>

2. 安装proto插件
    1. 安装protoc-gen-go(生成go代码)

> go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
>

    2. 安装protoc-gen-go-grpc(生成grpc代码)

> go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
>



3. 配置go环境变量

> export PATH="$PATH:$(go env GOPATH)/bin"
>

4. 验证插件可行性

> which protoc-gen-go       # 应输出 $GOPATH/bin/protoc-gen-go
>
> which protoc-gen-go-grpc  # 应输出 $GOPATH/bin/protoc-gen-go-grpc
>



5. 编写proto文件

```plain
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

6. 生成go文件以及grpc文件

> protoc --go_out=. --go-grpc_out=. hello.proto
>

<h1 id="Uu6X7">二、proto文件介绍</h1>
<h2 id="qMqyI">syntax</h2>
```plain
// 这里说明使用的是proto3语法
syntax = "proto3";
```

 

<h2 id="N4KSh">option go_package</h2>
```plain
// 这部分的内容是关于最后生成的go文件是放在哪个目录哪个包中，.代表在当前目录生成，service代表了生成的go文件的包名是service.
option go_package = ".;service";
```

 

<h2 id="ZAzgh">messgage关键字</h2>
+ message关键字类似于java中的class，go中的struct
+ 在消息中承载的数据分别对应每一个字段，其中每个字段都有一个名字和一种类型
+ 一个proto文件中可以定义多种消息类型

<h3 id="MN2qG">字段规则</h3>
+ required消息体中必填字段，不设置会导致编码异常。在proto2中使用，在proto3中剔除
+ optioanl消息体中可选字段，protobuf3没有required,optional等说明关键字，都默认为optional
+ repeate消息体中可重复字段，重复的值的顺序会被保留在go中重复的会被定义为切片。

<h3 id="BFt1f">消息号</h3>
在消息体的定义中，每个字段都必须要有一个唯一的标识号，一个整数。

<h3 id="SrjyT">嵌套消息</h3>
可以在其他消息类型中定义，使用消息类型，如下,person消息定义在PersonInfo消息内

```go
message PersonInfo {

    message Person {
        string name = 1;
        int32 height =2;
        repeated int32 weight = 3;
    } 
    repeated Person info = 1;
}
```

如果要在它的父消息类型外部重用这个消息，需要PersonInfo.Person的形式使用它，如：

```go
message PersonMessage {
    PersonInfo.Person info = 1;
}
```

<h2 id="p6Kdh">服务定义</h2>
如果想要将消息类型用在RPC系统中，可以在.proto文件中定义一个RPC服务接口。protocol buffer编译器会根据所选择的不同语言生成服务接口代码以及存根

```go
// 然后我们需要定义一个服务，在这个服务中需要有一个方法，这个方法可以接受客户端的参数，再返回服务响应。
//其实很容易可以看出，我们定义了一个service，称为SayHello,这个服务有一个rpc方法，名为SayHello
//这个方法会发送一个HelloRequest,然后返回一个HelloResponse
service SayHello{
  rpc SayHello(HelloRequest) returns(HelloResponse) {}
}

```



<h1 id="qDsFr">三、编写服务端代码</h1>
```go
package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "grpc-study/hello-server/proto"
	"net"
)

type server struct {
	pb.UnimplementedSayHelloServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	fmt.Printf("SayHello called with req: %v\n", req)
	return &pb.HelloResponse{ResponseMsg: "hello " + req.RequestName}, nil
}

func main() {
	// 开启端口
	listen, _ := net.Listen("tcp", ":8080")

	//创建grpc服务
	grpcServer := grpc.NewServer()

	// 在grpc服务端中注册我们自己编写的服务
	pb.RegisterSayHelloServer(grpcServer, &server{})

	//启动服务
	err := grpcServer.Serve(listen)

	if err != nil {
		panic(err)
	}
}

```

主要是实现hello_grpc_pb.go文件中的SayHelloServer接口

```go
type SayHelloServer interface {
	SayHello(context.Context, *HelloRequest) (*HelloResponse, error)
	mustEmbedUnimplementedSayHelloServer()
}

func (UnimplementedSayHelloServer) SayHello(context.Context, *HelloRequest) (*HelloResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SayHello not implemented")
}
```

<h1 id="DLwhH">四、编写客户端代码</h1>
```go
package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	service "grpc-study/hello-server/proto"
	"log"
	"time"
)

func main() {
	// 1. 创建gRpc客户端连接(推荐全局复用)
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials((insecure.NewCredentials())))
	if err != nil {
		log.Fatal("failed to create grpc client: %v", err)
	}
	defer conn.Close()

	// 2. 创建客户端
	c := service.NewSayHelloClient(conn)

	// 3. 调用服务端方法（使用context设置超时）
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &service.HelloRequest{
		RequestName: "张三",
	})
	if err != nil {
		log.Fatalf("clould not greet: %v", err)
	}
	log.Printf("服务端响应: %v", r.ResponseMsg)
}

```

<h1 id="tlxWD">五 调试</h1>
1. 启动服务端
2. 启动客户端

客户端控制台：

> 2025/07/17 11:15:45 服务端响应: hello 张三
>

服务端控制台：

> SayHello called with req: requestName:"张三"
>

