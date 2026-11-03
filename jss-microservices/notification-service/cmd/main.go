package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	mq "github.com/linhhuynhcoding/jss-microservices/mq"
	mqconfig "github.com/linhhuynhcoding/jss-microservices/mq/config"
	"github.com/linhhuynhcoding/jss-microservices/mq/events"
	"github.com/linhhuynhcoding/jss-microservices/notification-service/internal/grpc/server"
	"github.com/linhhuynhcoding/jss-microservices/notification-service/internal/repository"
	"github.com/linhhuynhcoding/jss-microservices/notification-service/internal/service"
	"github.com/linhhuynhcoding/jss-microservices/notification-service/pkg/config"
	"github.com/linhhuynhcoding/jss-microservices/notification-service/pkg/database"
	"github.com/linhhuynhcoding/jss-microservices/notification-service/pkg/logger"
	notificationpb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/notification"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func main() {
	// 1) Config + logger
	cfg := config.Load()
	logg := logger.New(cfg.LogLevel)
	defer logg.Sync()

	// 2) DB + repo + service
	db := database.Connect(cfg, logg)
	repo := repository.NewNotificationRepository(db, logg)
	svc := service.NewNotificationService(repo, logg)

	// 3) MQ subscriber (dùng /mq) — handler ký tên func([]byte) error
	var sub *mq.Subscriber
	if cfg.RabbitMQURL != "" {
		subscriberName := cfg.SubscriberName
		if subscriberName == "" {
			subscriberName = "noti.notification.create"
		}
		keys := cfg.BindingKeys
		if len(keys) == 0 {
			keys = []string{"notification.create"}
		}

		subCfg := mqconfig.RabbitMQConfig{
			ConnStr:        cfg.RabbitMQURL,
			ExchangeName:   cfg.ExchangeName,    // "notifications"
			ExchangeType:   "topic",
			SubscriberName: cfg.SubscriberName,  // "noti.notification.create"
			SubscribeKeys:  cfg.BindingKeys,    
		}	


		var err error
		sub, err = mq.NewSubscriber(subCfg, logg)
		if err != nil {
			logg.Fatal("mq subscriber init failed", zap.Error(err))
		}

		// Consume nhận []byte (envelope), tự giải mã rồi xử lý
		go func() {
			err := sub.Consume(func(b []byte) error {
			var env events.EventEnvelope
			if err := proto.Unmarshal(b, &env); err != nil {
					logg.Warn("drop: unmarshal envelope failed", zap.Error(err))
					return nil // ACK: lỗi không phục hồi
			}

			var msg notificationpb.CreateNotificationRequest
			if err := proto.Unmarshal(env.Payload, &msg); err != nil {
					logg.Warn("drop: unmarshal payload failed", zap.Error(err))
					return nil // ACK
			}

			// GỌI BUSINESS
			_, err := svc.Create(context.Background(), msg.UserId, "", msg.Title, msg.Message)
			if err == nil {
					return nil // ACK
			}

			low := strings.ToLower(err.Error())

			// ❌ LỖI KHÔNG PHỤC HỒI → ACK (KHÔNG requeue)
			if strings.Contains(low, "invalid user id") ||
				strings.Contains(low, "invalid id") ||
				strings.Contains(low, "validation") ||
				strings.Contains(low, "duplicate key") || // ví dụ DuplicateKey Mongo
				strings.Contains(low, "already exists") ||
				strings.Contains(low, "bad request") ||
				strings.Contains(low, "parse") {
					logg.Warn("drop: permanent error", zap.String("userId", msg.UserId), zap.Error(err))
					return nil // ACK & drop
			}

    // ✅ LỖI TẠM THỜI → TRẢ ERR để requeue (tạm thời)
    logg.Error("create notification failed (will requeue)", zap.Error(err))
    return err
			})

			if err != nil {
				logg.Fatal("mq subscriber exited", zap.Error(err))
			}
		}()

		logg.Info("rabbitmq subscriber started",
			zap.String("exchange", subCfg.ExchangeName),
			zap.String("subscriber", subCfg.SubscriberName),
			zap.Strings("keys", subCfg.SubscribeKeys),
		)
	}

	// 4) gRPC server
	grpcServer := grpc.NewServer()
	notificationpb.RegisterNotificationServiceServer(
		grpcServer,
		server.NewServer(svc, logg),
	)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logg.Fatal("failed to listen for gRPC", zap.Error(err))
	}
	go func() {
		logg.Info("gRPC listening", zap.String("port", cfg.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			logg.Fatal("gRPC serve failed", zap.Error(err))
		}
	}()

	// 5) gRPC-Gateway (REST)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		if err := notificationpb.RegisterNotificationServiceHandlerFromEndpoint(
			ctx, mux, "localhost:"+cfg.GRPCPort, opts,
		); err != nil {
			log.Fatal("failed to register notification gateway", err)
		}
		log.Printf("HTTP REST gateway listening on :%s", cfg.HTTPPort)
		if err := http.ListenAndServe(":"+cfg.HTTPPort, mux); err != nil {
			log.Fatal("HTTP REST serve failed:", err)
		}
	}()

	// 6) Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logg.Info("shutting down gRPC")
	grpcServer.GracefulStop()
	if sub != nil {
		sub.Close() // không có return value
	}
	logg.Info("exited")
}
