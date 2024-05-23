package grpcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"emperror.dev/errors"
	"github.com/go-thread-7/commonlib/grpc/grpc-server/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const (
	maxConnectionIdle = 5
	gRPCTimeout       = 15
	maxConnectionAge  = 5
	gRPCTime          = 10
)

type GrpcServer struct {
	Grpc   *grpc.Server
	Config *config.GRPCConfig
}

func New(config *config.GRPCConfig) *GrpcServer {
	s := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: maxConnectionIdle * time.Minute,
			Timeout:           gRPCTimeout * time.Second,
			MaxConnectionAge:  maxConnectionAge * time.Minute,
			Time:              gRPCTime * time.Minute,
		}),
	)

	return &GrpcServer{
		Grpc:   s,
		Config: config,
	}
}

func (s *GrpcServer) RunGrpcServer(ctx context.Context, configGrpc ...func(grpcServer *grpc.Server)) error {
	listen, err := net.Listen("tcp", s.Config.Port)
	if err != nil {
		return errors.Wrap(err, "net.Listen")
	}

	if len(configGrpc) > 0 {
		grpcFunc := configGrpc[0]
		if grpcFunc != nil {
			grpcFunc(s.Grpc)
		}
	}

	if s.Config.Development {
		reflection.Register(s.Grpc)
	}

	if len(configGrpc) > 0 {
		grpcFunc := configGrpc[0]
		if grpcFunc != nil {
			grpcFunc(s.Grpc)
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("shutting down grpc port: %s\n", s.Config.Port)
				s.shutdown()
				fmt.Println("grpc exited properly")
				return
			}
		}
	}()

	fmt.Printf("grpc server is listening on port: %s\n", s.Config.Port)

	err = s.Grpc.Serve(listen)

	if err != nil {
		fmt.Println(fmt.Sprintf("[grpcServer_RunGrpcServer.Serve] grpc server serve error: %+v", err))
	}

	return err
}

func (s *GrpcServer) shutdown() {
	s.Grpc.Stop()
	s.Grpc.GracefulStop()
}
