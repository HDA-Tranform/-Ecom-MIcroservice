package service

import (
	utils "github.com/linhhuynhcoding/jss-microservices/jss-shared/utils/format"
	db "github.com/linhhuynhcoding/jss-microservices/market/internal/repository"
	api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/market"
)

func (s *Service) dbGoldPrice2PbGoldPrice(dbGp *db.GoldPrice) *api.GoldPrice {
	return &api.GoldPrice{
		Id:        int64(dbGp.GoldID),
		GoldType:  dbGp.GoldType,
		BuyPrice:  float32(utils.NumericToFloat64(dbGp.BuyPrice)),
		SellPrice: float32(utils.NumericToFloat64(dbGp.SellPrice)),
		Date:      utils.PgToPbTimestamp(dbGp.Date),
	}
}
