package service

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/jackc/pgx/v5/pgtype"
	utils "github.com/linhhuynhcoding/jss-microservices/jss-shared/utils/format"
	db "github.com/linhhuynhcoding/jss-microservices/market/internal/repository"
	api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/market"
)

// CreateBuybackPolicy creates a new buyback policy record.
func (s *Service) CreateBuybackPolicy(ctx context.Context, req *api.CreateBuybackPolicyRequest) (*api.CreateBuybackPolicyResponse, error) {
	s.logger.Info("CreateBuybackPolicy called", zap.Any("req", req))

	arg := db.CreateBuybackPolicyParams{
		ProductType: req.ProductType,
		BuybackRate: utils.ToNumeric(float64(req.BuybackRate)),
		Description: utils.StringToPgText(req.Description),
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	bp, err := s.queries.CreateBuybackPolicy(ctx, arg)
	if err != nil {
		s.logger.Error("failed to create buyback policy", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create buyback policy: %v", err)
	}

	s.logger.Info("Buyback policy created successfully", zap.Any("buyback_policy", bp))
	return &api.CreateBuybackPolicyResponse{
		BuybackPolicy: s.dbBuybackPolicy2PbBuybackPolicy(&bp),
	}, nil
}

// GetBuybackPolicy retrieves a buyback policy by ID.
func (s *Service) GetBuybackPolicy(ctx context.Context, req *api.GetBuybackPolicyRequest) (*api.GetBuybackPolicyResponse, error) {
	s.logger.Info("GetBuybackPolicy called", zap.Any("req", req))

	bp, err := s.queries.GetBuybackPolicy(ctx, int32(req.Id))
	if err != nil {
		s.logger.Error("buyback policy not found", zap.Error(err), zap.Int64("id", req.Id))
		return nil, status.Errorf(codes.NotFound, "buyback policy not found: %v", err)
	}

	s.logger.Info("Buyback policy retrieved", zap.Any("buyback_policy", bp))
	return &api.GetBuybackPolicyResponse{
		BuybackPolicy: s.dbBuybackPolicy2PbBuybackPolicy(&bp),
	}, nil
}

// ListBuybackPolicies lists buyback policies with pagination and optional product_type filtering.
func (s *Service) ListBuybackPolicies(ctx context.Context, req *api.ListBuybackPoliciesRequest) (*api.ListBuybackPoliciesResponse, error) {
	s.logger.Info("ListBuybackPolicies called", zap.Any("req", req))

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	var bps []db.BuybackPolicy
	var totalCount int64
	var err error

	if req.ProductType != "" {
		// Filter by product type
		arg := db.GetBuybackPoliciesByProductTypeParams{
			ProductType: req.ProductType,
			Limit:       limit,
			Offset:      req.Offset,
		}
		bps, err = s.queries.GetBuybackPoliciesByProductType(ctx, arg)
		if err != nil {
			s.logger.Error("failed to list buyback policies by product type", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to list buyback policies: %v", err)
		}

		totalCount, err = s.queries.CountBuybackPoliciesByProductType(ctx, req.ProductType)
		if err != nil {
			s.logger.Error("failed to count buyback policies by product type", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to count buyback policies: %v", err)
		}
	} else {
		// No filter — return all
		arg := db.GetBuybackPoliciesParams{
			Limit:  limit,
			Offset: req.Offset,
		}
		bps, err = s.queries.GetBuybackPolicies(ctx, arg)
		if err != nil {
			s.logger.Error("failed to list buyback policies", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to list buyback policies: %v", err)
		}

		totalCount, err = s.queries.CountBuybackPolicies(ctx)
		if err != nil {
			s.logger.Error("failed to count buyback policies", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to count buyback policies: %v", err)
		}
	}

	var results []*api.BuybackPolicy
	for _, bp := range bps {
		results = append(results, s.dbBuybackPolicy2PbBuybackPolicy(&bp))
	}

	s.logger.Info("Buyback policies listed", zap.Int("count", len(results)))
	return &api.ListBuybackPoliciesResponse{
		BuybackPolicies: results,
		TotalCount:      int32(totalCount),
	}, nil
}

// GetBuybackPoliciesByProductType retrieves buyback policies filtered by product type with pagination.
func (s *Service) GetBuybackPoliciesByProductType(ctx context.Context, productType string, limit, offset int32) ([]*api.BuybackPolicy, error) {
	s.logger.Info("GetBuybackPoliciesByProductType called",
		zap.String("product_type", productType),
		zap.Int32("limit", limit),
		zap.Int32("offset", offset),
	)

	arg := db.GetBuybackPoliciesByProductTypeParams{
		ProductType: productType,
		Limit:       limit,
		Offset:      offset,
	}

	bps, err := s.queries.GetBuybackPoliciesByProductType(ctx, arg)
	if err != nil {
		s.logger.Error("failed to get buyback policies by product type", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get buyback policies by product type: %v", err)
	}

	var results []*api.BuybackPolicy
	for _, bp := range bps {
		results = append(results, s.dbBuybackPolicy2PbBuybackPolicy(&bp))
	}

	s.logger.Info("Buyback policies by product type retrieved", zap.Int("count", len(results)))
	return results, nil
}

// GetActiveBuybackPolicies retrieves buyback policies that have a positive buyback rate with pagination.
func (s *Service) GetActiveBuybackPolicies(ctx context.Context, limit, offset int32) ([]*api.BuybackPolicy, error) {
	s.logger.Info("GetActiveBuybackPolicies called",
		zap.Int32("limit", limit),
		zap.Int32("offset", offset),
	)

	arg := db.GetActiveBuybackPoliciesParams{
		Limit:  limit,
		Offset: offset,
	}

	bps, err := s.queries.GetActiveBuybackPolicies(ctx, arg)
	if err != nil {
		s.logger.Error("failed to get active buyback policies", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get active buyback policies: %v", err)
	}

	var results []*api.BuybackPolicy
	for _, bp := range bps {
		results = append(results, s.dbBuybackPolicy2PbBuybackPolicy(&bp))
	}

	s.logger.Info("Active buyback policies retrieved", zap.Int("count", len(results)))
	return results, nil
}

// UpdateBuybackPolicy updates an existing buyback policy.
func (s *Service) UpdateBuybackPolicy(ctx context.Context, req *api.UpdateBuybackPolicyRequest) (*api.UpdateBuybackPolicyResponse, error) {
	s.logger.Info("UpdateBuybackPolicy called", zap.Any("req", req))

	// Fetch existing record to preserve created_at
	existing, err := s.queries.GetBuybackPolicy(ctx, int32(req.Id))
	if err != nil {
		s.logger.Error("buyback policy not found for update", zap.Error(err), zap.Int64("id", req.Id))
		return nil, status.Errorf(codes.NotFound, "buyback policy not found: %v", err)
	}

	arg := db.UpdateBuybackPolicyParams{
		ID:          int32(req.Id),
		ProductType: req.ProductType,
		BuybackRate: utils.ToNumeric(float64(req.BuybackRate)),
		Description: utils.StringToPgText(req.Description),
		CreatedAt:   existing.CreatedAt,
	}

	bp, err := s.queries.UpdateBuybackPolicy(ctx, arg)
	if err != nil {
		s.logger.Error("failed to update buyback policy", zap.Error(err), zap.Int64("id", req.Id))
		return nil, status.Errorf(codes.Internal, "failed to update buyback policy: %v", err)
	}

	s.logger.Info("Buyback policy updated", zap.Any("buyback_policy", bp))
	return &api.UpdateBuybackPolicyResponse{
		BuybackPolicy: s.dbBuybackPolicy2PbBuybackPolicy(&bp),
	}, nil
}

// UpdateBuybackRate updates only the buyback_rate field of an existing policy.
func (s *Service) UpdateBuybackRate(ctx context.Context, id int32, rate float32) (*api.BuybackPolicy, error) {
	s.logger.Info("UpdateBuybackRate called", zap.Int32("id", id), zap.Float32("rate", rate))

	arg := db.UpdateBuybackRateParams{
		ID:          id,
		BuybackRate: utils.ToNumeric(float64(rate)),
	}

	bp, err := s.queries.UpdateBuybackRate(ctx, arg)
	if err != nil {
		s.logger.Error("failed to update buyback rate", zap.Error(err), zap.Int32("id", id))
		return nil, status.Errorf(codes.Internal, "failed to update buyback rate: %v", err)
	}

	s.logger.Info("Buyback rate updated", zap.Any("buyback_policy", bp))
	return s.dbBuybackPolicy2PbBuybackPolicy(&bp), nil
}

// DeleteBuybackPolicy deletes a buyback policy by ID.
func (s *Service) DeleteBuybackPolicy(ctx context.Context, req *api.DeleteBuybackPolicyRequest) (*emptypb.Empty, error) {
	s.logger.Info("DeleteBuybackPolicy called", zap.Any("req", req))

	err := s.queries.DeleteBuybackPolicy(ctx, int32(req.Id))
	if err != nil {
		s.logger.Error("failed to delete buyback policy", zap.Error(err), zap.Int64("id", req.Id))
		return nil, status.Errorf(codes.Internal, "failed to delete buyback policy: %v", err)
	}

	s.logger.Info("Buyback policy deleted", zap.Int64("id", req.Id))
	return &emptypb.Empty{}, nil
}

// CountBuybackPolicies returns the total number of buyback policy records.
func (s *Service) CountBuybackPolicies(ctx context.Context) (int64, error) {
	s.logger.Info("CountBuybackPolicies called")

	count, err := s.queries.CountBuybackPolicies(ctx)
	if err != nil {
		s.logger.Error("failed to count buyback policies", zap.Error(err))
		return 0, status.Errorf(codes.Internal, "failed to count buyback policies: %v", err)
	}

	s.logger.Info("Buyback policies counted", zap.Int64("count", count))
	return count, nil
}

// SearchBuybackPoliciesByProductType performs a fuzzy search on product_type with pagination.
func (s *Service) SearchBuybackPoliciesByProductType(ctx context.Context, keyword string, limit, offset int32) ([]*api.BuybackPolicy, error) {
	s.logger.Info("SearchBuybackPoliciesByProductType called",
		zap.String("keyword", keyword),
		zap.Int32("limit", limit),
		zap.Int32("offset", offset),
	)

	arg := db.SearchBuybackPoliciesByProductTypeParams{
		Column1: pgtype.Text{String: keyword, Valid: true},
		Limit:   limit,
		Offset:  offset,
	}

	bps, err := s.queries.SearchBuybackPoliciesByProductType(ctx, arg)
	if err != nil {
		s.logger.Error("failed to search buyback policies", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to search buyback policies: %v", err)
	}

	var results []*api.BuybackPolicy
	for _, bp := range bps {
		results = append(results, s.dbBuybackPolicy2PbBuybackPolicy(&bp))
	}

	s.logger.Info("Buyback policies search completed", zap.Int("count", len(results)))
	return results, nil
}
