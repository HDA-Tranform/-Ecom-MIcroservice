package market_service

import (
	"context"

	"github.com/linhhuynhcoding/jss-microservices/product/config"
	"github.com/linhhuynhcoding/jss-microservices/product/pkg/grpc"
	"go.uber.org/zap"

	api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/market"
)

type IMarketServiceClient interface {
	GetGoldPrice(ctx context.Context, req *api.GetGoldPriceRequest) (*api.GetGoldPriceResponse, error)
}

type MarketServiceClient struct {
	client api.MarketClient
	logger *zap.Logger
	cfg    config.Config
}

func NewMarketServiceClient(
	logger *zap.Logger,
	cfg config.Config,
) IMarketServiceClient {
	return &MarketServiceClient{
		logger: logger,
		cfg:    cfg,
	}
}

func (m *MarketServiceClient) Connect() error {
	if m.client == nil {
		conn, err := grpc.NewConnection(m.cfg.MarketServiceUrl)
		if err != nil {
			m.logger.Error("failed to connect to market service", zap.Error(err))
			return err
		}
		m.client = api.NewMarketClient(conn)
	}
	return nil
}

func (m *MarketServiceClient) GetGoldPrice(ctx context.Context, req *api.GetGoldPriceRequest) (*api.GetGoldPriceResponse, error) {
	log := m.logger.With(zap.String("func", "GetGoldPrice"))
	if err := m.Connect(); err != nil {
		log.Error("failed to connect to market service", zap.Error(err))
		return nil, err
	}

	resp, err := m.client.GetGoldPrice(ctx, req)
	if err != nil {
		log.Error("failed to get gold price", zap.Error(err))
		return nil, err
	}
	return resp, err
}
