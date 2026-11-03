package main

// Entry point for the order service.  It sets up configuration, logging,
// database connections and external service clients, then starts both the
// gRPC server and its accompanying gRPC‑Gateway HTTP server.  When
// running under Docker this binary is invoked by the order-service
// container.

import (
    "context"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "github.com/linhhuynhcoding/jss-microservices/order-service/config"
    "github.com/linhhuynhcoding/jss-microservices/order-service/internal/service"
    orderpb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/order"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Load configuration from environment with sensible defaults
    cfg := config.Load()

    // Initialise logger
    logger, err := zap.NewProduction()
    if err != nil {
        log.Fatalf("failed to initialise logger: %v", err)
    }
    defer logger.Sync()

    // Connect to MongoDB
    mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
    if err != nil {
        logger.Fatal("failed to connect to MongoDB", zap.Error(err))
    }
    db := mongoClient.Database(cfg.MongoDB)

    // Initialise order service
    orderService, err := service.New(cfg, db, logger)
    if err != nil {
        logger.Fatal("failed to initialise order service", zap.Error(err))
    }
    defer orderService.Close()

    // Start gRPC server
    grpcServer := grpc.NewServer()
    orderpb.RegisterOrderServiceServer(grpcServer, orderService)
    grpcLis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        logger.Fatal("failed to listen for gRPC", zap.Error(err))
    }
    go func() {
        logger.Info("gRPC listening", zap.String("port", cfg.GRPCPort))
        if err := grpcServer.Serve(grpcLis); err != nil {
            logger.Fatal("gRPC server terminated", zap.Error(err))
        }
    }()

    // Start HTTP server via gRPC‑Gateway
    go func() {
        ctx := context.Background()
        mux := runtime.NewServeMux()
        opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
        endpoint := fmt.Sprintf("localhost:%s", cfg.GRPCPort)
        if err := orderpb.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
            logger.Fatal("failed to start HTTP gateway", zap.Error(err))
        }
        logger.Info("HTTP gateway listening", zap.String("port", cfg.HTTPPort))
        if err := http.ListenAndServe(":"+cfg.HTTPPort, mux); err != nil {
            logger.Fatal("HTTP server terminated", zap.Error(err))
        }
    }()

    // Graceful shutdown on interrupt or termination
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    logger.Info("shutting down order service")
    grpcServer.GracefulStop()
    // Close external connections
    _ = mongoClient.Disconnect(context.Background())
}