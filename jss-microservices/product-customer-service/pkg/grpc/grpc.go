package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewConnection(url string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithChainUnaryInterceptor())

	return conn, err
}
