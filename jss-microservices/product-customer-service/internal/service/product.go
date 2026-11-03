package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/linhhuynhcoding/jss-microservices/product/consts"
	db "github.com/linhhuynhcoding/jss-microservices/product/internal/repository"
	utils "github.com/linhhuynhcoding/jss-microservices/product/utils/numeric"
	market_api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/market"
	api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/product"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) CreateProduct(ctx context.Context, req *api.CreateProductRequest) (*api.ProductResponse, error) {
	log := s.logger.With(zap.String("func", "CreateProduct"))
	log.Info("req", zap.Any("req", req))

	goldPrice, err := s.adapter.marketClient.GetGoldPrice(ctx, &market_api.GetGoldPriceRequest{
		Id: int64(req.GoldType),
	})
	if err != nil {
		log.Error("cannot get gold price", zap.Error(err))
		return nil, fmt.Errorf("cannot get gold price")
	}
	goldBuyPrice := goldPrice.GoldPrice.BuyPrice

	// Giá bán = giá vốn sản phẩm * tỉ lệ áp giá,
	// Giá vốn sản phẩm = [giá vàng thời điểm * trọng lượng sản phẩm] + tiền công + tiền đá
	sellingPrice := (1 + req.MarkupRate) * (float64(goldBuyPrice)*req.Weight/consts.MACE_OF_GOLD_WEIGHT + req.LaborCost + req.StoneCost)

	arg := db.CreateProductParams{
		Name:            pgtype.Text{String: req.Name, Valid: true},
		Code:            req.Code,
		CategoryID:      utils.Int32(req.CategoryId),
		Weight:          utils.ToNumeric(req.Weight),
		LaborCost:       utils.ToNumeric(req.LaborCost),
		StoneCost:       utils.ToNumeric(req.StoneCost),
		MarkupRate:      utils.ToNumeric(req.MarkupRate),
		SellingPrice:    utils.ToNumeric(sellingPrice),
		WarrantyPeriod:  utils.Int32(req.WarrantyPeriod),
		Stock:           utils.Int32(req.Stock),
		Image:           pgtype.Text{String: req.Image, Valid: true},
		GoldPriceAtTime: utils.ToNumeric(float64(goldPrice.GoldPrice.BuyPrice)),
	}
	log.Info("args", zap.Any("args", arg))

	product, err := s.queries.CreateProduct(ctx, arg)
	if err != nil {
		log.Error("failed to create product", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return &api.ProductResponse{Product: s.productToProto(product)}, nil
}

func (s *Service) GetProduct(ctx context.Context, req *api.GetProductRequest) (*api.ProductResponse, error) {
	product, err := s.queries.GetProductByID(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}

	return &api.ProductResponse{Product: s.productToProto(product)}, nil
}

func (s *Service) ListProducts(ctx context.Context, req *api.ListProductsRequest) (*api.ListProductsResponse, error) {
	products, err := s.queries.ListProducts(ctx, db.ListProductsParams{
		Limit:  req.Limit,
		Offset: req.Page * req.Limit,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	resp := &api.ListProductsResponse{}
	for _, p := range products {
		resp.Products = append(resp.Products, s.productToProto(p))
	}
	return resp, nil
}

// TODO: tính lại giá
func (s *Service) UpdateProduct(ctx context.Context, req *api.UpdateProductRequest) (*api.ProductResponse, error) {
	arg := db.UpdateProductByCodeParams{
		Name:           pgtype.Text{String: req.Name, Valid: true},
		Code:           req.Code,
		CategoryID:     utils.Int32(req.CategoryId),
		Weight:         utils.ToNumeric(req.Weight),
		LaborCost:      utils.ToNumeric(req.LaborCost),
		StoneCost:      utils.ToNumeric(req.StoneCost),
		MarkupRate:     utils.ToNumeric(req.MarkupRate),
		WarrantyPeriod: utils.Int32(req.WarrantyPeriod),
		Stock:          utils.Int32(req.Stock),
		Image:          pgtype.Text{String: req.Image, Valid: true},
	}

	product, err := s.queries.UpdateProductByCode(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	return &api.ProductResponse{Product: s.productToProto(product)}, nil
}

func (s *Service) DeleteProduct(ctx context.Context, req *api.DeleteProductRequest) (*api.DeleteProductResponse, error) {
	err := s.queries.DeleteProduct(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	return &api.DeleteProductResponse{Success: true}, nil
}

func (s *Service) ListProductCategories(ctx context.Context, req *api.ListProductCategoriesRequest) (*api.ListProductCategoriesResponse, error) {
	// Call the sqlc query
	categories, err := s.queries.ListProductCategories(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list categories: %v", err)
	}

	// Map db models to API models
	var res []*api.ProductCategory
	for _, c := range categories {
		res = append(res, &api.ProductCategory{
			Id:   int32(c.ID),
			Name: c.Name,
		})
	}

	return &api.ListProductCategoriesResponse{
		Categories: res,
	}, nil
}

func (s *Service) PurchaseProduct(ctx context.Context, req *api.PurchaseProductRequest) (*api.PurchaseProductResponse, error) {
	log := s.logger.With(zap.String("func", "PurchaseProduct"))
	log.Info("req", zap.Any("req", req))

	if len(req.Products) == 0 {
		log.Error("no products to purchase")
		return nil, status.Error(codes.InvalidArgument, "no products to purchase")
	}

	if req.OrderId == 0 {
		log.Error("invalid order id")
		return nil, status.Error(codes.InvalidArgument, "invalid order id")
	}

	customer, err := s.queries.GetCustomerByID(ctx, req.CustomerId)
	if err != nil {
		log.Error("failed to get customer", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get customer")
	}
	log.Info("customer", zap.Any("customer", customer))

	productIds := make([]int32, 0)
	for _, p := range req.Products {
		productIds = append(productIds, p.ProductId)
	}

	products, err := s.queries.GetProductsById(ctx, productIds)
	if err != nil {
		log.Error("failed to get products", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get products")
	}
	if len(products) != len(req.Products) {
		log.Error("failed to get products", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get products")
	}
	mapProductId := make(map[int32]db.Product)
	for _, p := range products {
		mapProductId[int32(p.ID)] = p
	}

	err = s.queries.ExecTx(ctx, func(q *db.Queries) error {
		for _, p := range req.Products {
			product := mapProductId[p.ProductId]
			if product.Stock.Int32 < p.Quantity {
				log.Error("not enough stock", zap.Error(err))
				return status.Error(codes.Internal, "not enough stock")
			}
			product.Stock.Int32 -= p.Quantity
			_, err := q.UpdateProductByCode(ctx, db.UpdateProductByCodeParams{
				Code:    product.Code,
				Stock:   product.Stock,
				BuyTurn: utils.Int32(product.BuyTurn.Int32 + 1),
			})
			if err != nil {
				log.Error("failed to update product", zap.Error(err))
				return status.Error(codes.Internal, "failed to update product")
			}
		}
		return nil
	})
	if err != nil {
		log.Error("failed to update product", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update product")
	}

	for _, p := range req.Products {
		_, err := s.queries.CreateOrderRecord(ctx, db.CreateOrderRecordParams{
			CustomerID: req.CustomerId,
			OrderID:    req.OrderId,
			ProductID:  int32(p.ProductId),
			Quantity:   p.Quantity,
		})
		if err != nil {
			log.Error("failed to create order record", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to create order record")
		}
	}
	log.Info("products", zap.Any("products", products))

	productReps := make([]*api.Product, 0)
	for _, p := range products {
		productReps = append(productReps, s.productToProto(p))
	}

	return &api.PurchaseProductResponse{
		Products: productReps,
		Customer: s.mapCustomerToProto(customer),
	}, nil
}
