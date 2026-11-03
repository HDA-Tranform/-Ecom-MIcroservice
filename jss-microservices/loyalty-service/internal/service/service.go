package service

import (
	"context"

	"github.com/linhhuynhcoding/jss-microservices/loyalty/config"
	"github.com/linhhuynhcoding/jss-microservices/loyalty/internal/repository"
	"github.com/linhhuynhcoding/jss-microservices/rpc/gen/loyalty"
	"go.uber.org/zap"
)

type Service struct {
	loyalty.UnimplementedLoyaltyServer

	logger  *zap.Logger
	cfg     config.Config
	queries repository.Store
}

func NewService(ctx context.Context, logger *zap.Logger, cfg config.Config, store repository.Store) *Service {

	return &Service{
		logger:  logger,
		cfg:     cfg,
		queries: store,
	}
}
