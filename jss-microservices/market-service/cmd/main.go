package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linhhuynhcoding/jss-microservices/market/config"
	"github.com/linhhuynhcoding/jss-microservices/market/internal/handler"
	"github.com/linhhuynhcoding/jss-microservices/market/internal/repository"
	"github.com/linhhuynhcoding/jss-microservices/market/internal/service"
	"github.com/linhhuynhcoding/jss-microservices/rpc/gen/market"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: IMPLEMENT FX
func main() {
	// ------------------------------------------------------------
	// 		INIT VARIABLES
	// ------------------------------------------------------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()
	log, _ := zap.NewProduction()
	defer log.Sync()
	log.Info("Config", zap.Any("cfg", cfg))

	// Determine ports (fallback to defaults if not configured)
	grpcPort := cfg.GrpcPort
	if grpcPort == 0 {
		grpcPort = 50051
	}
	httpPort := cfg.HttpPort
	if httpPort == 0 {
		httpPort = 8080
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	httpAddr := fmt.Sprintf(":%d", httpPort)

	// ------------------------------------------------------------
	// 		INIT DB
	// ------------------------------------------------------------
	connPool, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", zap.Error(err))
	}
	defer connPool.Close()
	store := repository.NewStore(connPool)

	{
		goldPriceCrawler := handler.NewGoldPriceCrawler(ctx, log, cfg, store)
		go RunCronJob(ctx, log, "0 */10 * * * *", goldPriceCrawler.Handle)
	}
	{
		go NewServer(ctx, cfg, log, store, grpcAddr)
	}
	{
		go NewGatewayServer(ctx, cfg, log, grpcAddr, httpAddr)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down market-service...")
}

func NewServer(
	ctx context.Context,
	cfg config.Config,
	log *zap.Logger,
	store repository.Store,
	grpcAddr string,
) {
	// ------------------------------------------------------------
	// 		START SERVER
	// ------------------------------------------------------------
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()
	market.RegisterMarketServer(s, service.NewService(ctx, log, cfg, store))

	log.Info("gRPC server listening", zap.String("addr", grpcAddr))
	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve", zap.Error(err))
	}
}

func NewGatewayServer(
	ctx context.Context,
	cfg config.Config,
	log *zap.Logger,
	grpcAddr string,
	httpAddr string,
) {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := market.RegisterMarketHandlerFromEndpoint(ctx, mux, "localhost"+grpcAddr, opts)
	if err != nil {
		log.Fatal("failed to start gateway", zap.Error(err))
	}

	log.Info("gRPC-Gateway listening", zap.String("addr", httpAddr))
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal("failed to serve gateway", zap.Error(err))
	}
}

func RunCronJob(
	ctx context.Context,
	logger *zap.Logger,
	schedule string,
	f func(context.Context) error,
) {
	// Cron
	c := cron.New(cron.WithSeconds()) // cho phép định nghĩa theo giây nếu muốn
	// Chạy job mỗi x phút
	_, err := c.AddFunc(schedule, func() {
		_ = f(ctx)
	})
	if err != nil {
		logger.Error("Failed to schedule cron", zap.Error(err))
	}

	c.Start()
	logger.Info("Cron job started.")

	// Giữ app chạy
	select {}
}
