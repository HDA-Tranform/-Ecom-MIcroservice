package service

import (
	utils "github.com/linhhuynhcoding/jss-microservices/jss-shared/utils/format"
	db "github.com/linhhuynhcoding/jss-microservices/market/internal/repository"
	api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/market"
)

func (s *Service) dbBuybackPolicy2PbBuybackPolicy(bp *db.BuybackPolicy) *api.BuybackPolicy {
	return &api.BuybackPolicy{
		Id:          int64(bp.ID),
		ProductType: bp.ProductType,
		BuybackRate: float32(utils.NumericToFloat64(bp.BuybackRate)),
		Description: utils.PgTextToString(bp.Description),
		CreatedAt:   utils.PgToPbTimestamp(bp.CreatedAt),
	}
}
