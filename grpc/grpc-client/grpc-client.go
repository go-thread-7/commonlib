package grpcclient

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcConfig struct {
	Port        string `mapstructure:"port"`
	Host        string `mapstructure:"host"`
	Development bool   `mapstructure:"development"`
}

type grpcClient struct {
	conn *grpc.ClientConn
}

type GrpcClient interface {
	GetGrpcConnection() *grpc.ClientConn
	Close() error
}

func New(config *GrpcConfig) (GrpcClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s%s", config.Host, config.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &grpcClient{conn: conn}, nil
}

func (g *grpcClient) GetGrpcConnection() *grpc.ClientConn {
	return g.conn
}

func (g *grpcClient) Close() error {
	return g.conn.Close()
}
