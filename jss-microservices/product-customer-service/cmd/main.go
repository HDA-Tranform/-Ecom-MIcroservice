package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linhhuynhcoding/jss-microservices/product/config"
	"github.com/linhhuynhcoding/jss-microservices/product/internal/repository"
	"github.com/linhhuynhcoding/jss-microservices/product/internal/service"
	"github.com/linhhuynhcoding/jss-microservices/rpc/gen/product"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: IMPLEMENT FX
func main() {
	// ------------------------------------------------------------
	// 		INIT VARIABLES
	// ------------------------------------------------------------
	ctx := context.Background()
	log, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to init logger! %v", err)
		return
	}

	cfg := config.NewConfig()
	log.Info("Config: ", zap.Any("cfg", cfg))

	// ------------------------------------------------------------
	// 		INIT DB
	// ------------------------------------------------------------
	connPool, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db")
	}
	store := repository.NewStore(connPool)

	s := service.NewService(ctx, log, cfg, store)

	go NewServer(ctx, cfg, log, s)
	NewGatewayServer(ctx, cfg, log, s)
}

func NewServer(
	ctx context.Context,
	cfg config.Config,
	log *zap.Logger,
	service *service.Service,
) {

	// ------------------------------------------------------------
	// 		START SERVER
	// ------------------------------------------------------------
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.GrpcPort))
	if err != nil {
		log.Fatal("failed to listen: %v", zap.Error(err))
	}

	s := grpc.NewServer()
	product.RegisterProductCustomerServer(s, service)

	log.Info("gRPC server listening", zap.Any("port", cfg.GrpcPort))
	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve: %v", zap.Error(err))
	}
}

func NewGatewayServer(
	ctx context.Context,
	cfg config.Config,
	log *zap.Logger,
	service *service.Service,
) {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := product.RegisterProductCustomerHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%v", cfg.GrpcPort), opts)
	if err != nil {
		log.Fatal("failed to start gateway", zap.Error(err))
	}

	mux.HandlePath("POST", "/v1/upload", service.UploadFileHTTP)

	log.Info("gRPC-Gateway listening on :8080")
	if err := http.ListenAndServe(fmt.Sprintf(":%v", cfg.HttpPort), mux); err != nil {
		log.Fatal("failed to serve: %v", zap.Error(err))
	}
}
