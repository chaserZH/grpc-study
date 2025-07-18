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
