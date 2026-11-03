package service

import (
	"time"

	db "github.com/linhhuynhcoding/jss-microservices/product/internal/repository"
	utils "github.com/linhhuynhcoding/jss-microservices/product/utils/numeric"
	api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/product"
)

func (s *Service) productToProto(p db.Product) *api.Product {
	return &api.Product{
		Id:              p.ID,
		Name:            p.Name.String,
		Code:            p.Code,
		CategoryId:      p.CategoryID.Int32,
		Weight:          utils.NumericToFloat64(p.Weight),
		GoldPriceAtTime: utils.NumericToFloat64(p.GoldPriceAtTime),
		LaborCost:       utils.NumericToFloat64(p.LaborCost),
		StoneCost:       utils.NumericToFloat64(p.StoneCost),
		MarkupRate:      utils.NumericToFloat64(p.MarkupRate),
		SellingPrice:    utils.NumericToFloat64(p.SellingPrice),
		Stock:           p.Stock.Int32,
		BuyTurn:         p.BuyTurn.Int32,
		WarrantyPeriod:  p.WarrantyPeriod.Int32,
		Image:           p.Image.String,
		CreatedAt:       p.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:       p.UpdatedAt.Time.Format(time.RFC3339),
	}
}
