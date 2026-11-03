package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linhhuynhcoding/jss-microservices/loyalty/config"
	"github.com/linhhuynhcoding/jss-microservices/loyalty/internal/handler"
	"github.com/linhhuynhcoding/jss-microservices/loyalty/internal/repository"
	"github.com/linhhuynhcoding/jss-microservices/loyalty/internal/service"
	"github.com/linhhuynhcoding/jss-microservices/rpc/gen/loyalty"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// ------------------------------------------------------------
	//		INIT VARIABLES
	// ------------------------------------------------------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()
	log, _ := zap.NewProduction()
	defer log.Sync()

	log.Info("config", zap.Any("cfg", cfg))

	// ------------------------------------------------------------
	//		INIT DB
	// ------------------------------------------------------------
	connPool, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", zap.Error(err))
	}
	defer connPool.Close()
	store := repository.NewStore(connPool)

	// ------------------------------------------------------------
	//		INIT CONSUMERS
	// ------------------------------------------------------------
	{
		orderCreatedConsumer := handler.NewOrderCreatedConsumer(log, cfg, store)
		go orderCreatedConsumer.ConsumeOrderCreated(ctx)
	}

	// ------------------------------------------------------------
	//		START gRPC SERVER
	// ------------------------------------------------------------
	{
		go NewServer(ctx, cfg, log, store)
	}

	// ------------------------------------------------------------
	//		START HTTP GATEWAY
	// ------------------------------------------------------------
	{
		go NewGatewayServer(ctx, cfg, log)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down loyalty-service")
	cancel()
}

func NewServer(
	ctx context.Context,
	cfg config.Config,
	log *zap.Logger,
	store repository.Store,
) {
	grpcPort := cfg.GrpcPort
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()
	loyalty.RegisterLoyaltyServer(s, service.NewService(ctx, log, cfg, store))

	log.Info("gRPC server listening", zap.String("port", grpcPort))
	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve gRPC", zap.Error(err))
	}
}

func NewGatewayServer(
	ctx context.Context,
	cfg config.Config,
	log *zap.Logger,
) {
	httpPort := cfg.HttpPort
	if httpPort == "" {
		httpPort = "8080"
	}
	grpcPort := cfg.GrpcPort
	if grpcPort == "" {
		grpcPort = "50051"
	}

	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := loyalty.RegisterLoyaltyHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts)
	if err != nil {
		log.Fatal("failed to start gateway", zap.Error(err))
	}

	log.Info("gRPC-Gateway listening", zap.String("port", httpPort))
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		log.Fatal("failed to serve HTTP gateway", zap.Error(err))
	}
}
