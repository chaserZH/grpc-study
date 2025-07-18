# grpc-study
grpc入门学习


本个作为go语言学习中grpc入门教程
采用从无注册中心以及consul注册中心，完成grpc通信。

# 项目代码如下
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

# 必要准备环境
1、consul
2、proto插件包
> go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
> go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

具体的详细教程以及细节知识参考，doc文档下的教程
