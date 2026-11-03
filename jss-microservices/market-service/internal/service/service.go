package service

import (
	"context"

	"github.com/linhhuynhcoding/jss-microservices/market/config"
	"github.com/linhhuynhcoding/jss-microservices/market/internal/repository"
	"github.com/linhhuynhcoding/jss-microservices/rpc/gen/market"
	"go.uber.org/zap"
)

type Service struct {
	market.UnimplementedMarketServer

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
