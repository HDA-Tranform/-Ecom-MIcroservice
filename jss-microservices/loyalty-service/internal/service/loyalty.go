package service

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	utils "github.com/linhhuynhcoding/jss-microservices/jss-shared/utils/format"
	db "github.com/linhhuynhcoding/jss-microservices/loyalty/internal/repository"
	api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/loyalty"
)

// ===== LOYALTY POINTS HANDLERS =====

func (s *Service) CreateLoyaltyPoint(ctx context.Context, req *api.CreateLoyaltyPointRequest) (*api.GetLoyaltyPointResponse, error) {
	s.logger.Info("Creating loyalty point",
		zap.Int32("customer_id", req.CustomerId),
		zap.Int32("points", req.Points),
		zap.String("source", req.Source))

	if req.CustomerId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "customer_id must be positive")
	}
	if req.Source == "" {
		return nil, status.Error(codes.InvalidArgument, "source is required")
	}

	createParams := db.CreateLoyaltyPointParams{
		CustomerID: req.CustomerId,
		Points:     utils.Int32(req.Points),
		Source:     req.Source,
		// ReferenceID: ConvertOptionalInt32ToNullInt32(req.ReferenceId),
	}

	loyaltyPoint, err := s.queries.CreateLoyaltyPoint(ctx, createParams)
	if err != nil {
		s.logger.Error("Failed to create loyalty point", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create loyalty point")
	}

	response := &api.GetLoyaltyPointResponse{
		LoyaltyPoint: ConvertDBLoyaltyPointToAPI(loyaltyPoint),
	}

	s.logger.Info("Successfully created loyalty point", zap.Int32("id", loyaltyPoint.ID))
	return response, nil
}

func (s *Service) GetLoyaltyPoint(ctx context.Context, req *api.GetLoyaltyPointRequest) (*api.GetLoyaltyPointResponse, error) {
	s.logger.Info("Getting loyalty point", zap.Int32("id", req.Id))

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id must be positive")
	}

	loyaltyPoint, err := s.queries.GetLoyaltyPoint(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get loyalty point", zap.Error(err), zap.Int32("id", req.Id))
		return nil, status.Error(codes.NotFound, "loyalty point not found")
	}

	response := &api.GetLoyaltyPointResponse{
		LoyaltyPoint: ConvertDBLoyaltyPointToAPI(loyaltyPoint),
	}

	return response, nil
}

func (s *Service) GetLoyaltyPointsByCustomer(ctx context.Context, req *api.GetLoyaltyPointsByCustomerRequest) (*api.GetLoyaltyPointsResponse, error) {
	s.logger.Info("Getting loyalty points by customer", zap.Int32("customer_id", req.CustomerId))

	if req.CustomerId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "customer_id must be positive")
	}

	var (
		limit int32 = 10
		page  int32 = 1
	)

	if req.Pagination != nil {
		limit = req.Pagination.Limit
		page = req.Pagination.Page
	}

	params := db.GetLoyaltyPointsByCustomerParams{
		CustomerID: req.CustomerId,
		Limit:      limit,
		Offset:     (page - 1) * limit,
	}

	loyaltyPoints, err := s.queries.GetLoyaltyPointsByCustomer(ctx, params)
	if err != nil {
		s.logger.Error("Failed to get loyalty points by customer", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get loyalty points")
	}

	apiLoyaltyPoints := make([]*api.LoyaltyPoint, len(loyaltyPoints))
	for i, lp := range loyaltyPoints {
		apiLoyaltyPoints[i] = ConvertDBLoyaltyPointToAPI(lp)
	}

	response := &api.GetLoyaltyPointsResponse{
		LoyaltyPoints: apiLoyaltyPoints,
		Pagination: &api.PaginationResponse{
			Total:   int32(len(loyaltyPoints)), // This should ideally be a separate count query
			Limit:   limit,
			Page:    page,
			HasNext: len(loyaltyPoints) == int(limit),
		},
	}

	return response, nil
}

func (s *Service) GetLoyaltyPointsBySource(ctx context.Context, req *api.GetLoyaltyPointsBySourceRequest) (*api.GetLoyaltyPointsResponse, error) {
	s.logger.Info("Getting loyalty points by source", zap.String("source", req.Source))

	if req.Source == "" {
		return nil, status.Error(codes.InvalidArgument, "source is required")
	}

	var (
		limit int32 = 10
		page  int32 = 1
	)

	if req.Pagination != nil {
		limit = req.Pagination.Limit
		page = req.Pagination.Page
	}

	params := db.GetLoyaltyPointsBySourceParams{
		Source: req.Source,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	loyaltyPoints, err := s.queries.GetLoyaltyPointsBySource(ctx, params)
	if err != nil {
		s.logger.Error("Failed to get loyalty points by source", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get loyalty points")
	}

	apiLoyaltyPoints := make([]*api.LoyaltyPoint, len(loyaltyPoints))
	for i, lp := range loyaltyPoints {
		apiLoyaltyPoints[i] = ConvertDBLoyaltyPointToAPI(lp)
	}

	response := &api.GetLoyaltyPointsResponse{
		LoyaltyPoints: apiLoyaltyPoints,
		Pagination: &api.PaginationResponse{
			Total:   int32(len(loyaltyPoints)),
			Limit:   limit,
			Page:    page,
			HasNext: len(loyaltyPoints) == int(limit),
		},
	}

	return response, nil
}

func (s *Service) GetAllLoyaltyPoints(ctx context.Context, req *api.GetAllLoyaltyPointsRequest) (*api.GetLoyaltyPointsResponse, error) {
	s.logger.Info("Getting all loyalty points")

	var (
		limit int32 = 10
		page  int32 = 1
	)

	if req.Pagination != nil {
		limit = req.Pagination.Limit
		page = req.Pagination.Page
	}

	params := db.GetAllLoyaltyPointsParams{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	loyaltyPoints, err := s.queries.GetAllLoyaltyPoints(ctx, params)
	if err != nil {
		s.logger.Error("Failed to get all loyalty points", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get loyalty points")
	}

	apiLoyaltyPoints := make([]*api.LoyaltyPoint, len(loyaltyPoints))
	for i, lp := range loyaltyPoints {
		apiLoyaltyPoints[i] = ConvertDBLoyaltyPointToAPI(lp)
	}

	response := &api.GetLoyaltyPointsResponse{
		LoyaltyPoints: apiLoyaltyPoints,
		Pagination: &api.PaginationResponse{
			Total:   int32(len(loyaltyPoints)),
			Limit:   limit,
			Page:    page,
			HasNext: len(loyaltyPoints) == int(limit),
		},
	}

	return response, nil
}

func (s *Service) UpdateLoyaltyPoint(ctx context.Context, req *api.UpdateLoyaltyPointRequest) (*api.GetLoyaltyPointResponse, error) {
	s.logger.Info("Updating loyalty point", zap.Int32("id", req.Id))

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id must be positive")
	}
	if req.Source == "" {
		return nil, status.Error(codes.InvalidArgument, "source is required")
	}

	updateParams := db.UpdateLoyaltyPointsParams{
		ID:     req.Id,
		Points: utils.Int32(req.Points),
		Source: req.Source,
		// ReferenceID: ConvertOptionalInt32ToNullInt32(req.ReferenceId),
	}

	loyaltyPoint, err := s.queries.UpdateLoyaltyPoints(ctx, updateParams)
	if err != nil {
		s.logger.Error("Failed to update loyalty point", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update loyalty point")
	}

	response := &api.GetLoyaltyPointResponse{
		LoyaltyPoint: ConvertDBLoyaltyPointToAPI(loyaltyPoint),
	}

	s.logger.Info("Successfully updated loyalty point", zap.Int32("id", req.Id))
	return response, nil
}

func (s *Service) DeleteLoyaltyPoint(ctx context.Context, req *api.DeleteLoyaltyPointRequest) (*emptypb.Empty, error) {
	s.logger.Info("Deleting loyalty point", zap.Int32("id", req.Id))

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id must be positive")
	}

	err := s.queries.DeleteLoyaltyPoint(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to delete loyalty point", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete loyalty point")
	}

	s.logger.Info("Successfully deleted loyalty point", zap.Int32("id", req.Id))
	return &emptypb.Empty{}, nil
}

func (s *Service) GetCustomerTotalPoints(ctx context.Context, req *api.GetCustomerTotalPointsRequest) (*api.GetCustomerTotalPointsResponse, error) {
	s.logger.Info("Getting customer total points", zap.Int32("customer_id", req.CustomerId))

	if req.CustomerId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "customer_id must be positive")
	}

	total, err := s.queries.GetCustomerTotalPoints(ctx, req.CustomerId)
	if err != nil {
		s.logger.Error("Failed to get customer total points", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get customer total points")
	}

	// Convert interface{} to int32 - assuming it's a numeric value
	var totalPoints int32
	switch v := total.(type) {
	case int32:
		totalPoints = v
	case int64:
		totalPoints = int32(v)
	case int:
		totalPoints = int32(v)
	case float64:
		totalPoints = int32(v)
	default:
		s.logger.Error("Unexpected type for total points", zap.Any("total", total))
		return nil, status.Error(codes.Internal, "invalid total points format")
	}

	response := &api.GetCustomerTotalPointsResponse{
		TotalPoints: totalPoints,
		CustomerId:  req.CustomerId,
	}

	return response, nil
}
