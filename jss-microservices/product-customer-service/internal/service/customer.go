package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/linhhuynhcoding/jss-microservices/product/internal/repository"
	pb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/product"
)

// ----- CreateCustomer -----
func (s *Service) CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.CustomerResponse, error) {
	customer, err := s.queries.CreateCustomer(ctx, db.CreateCustomerParams{
		Name:    req.Name,
		Phone:   req.Phone,
		Email:   pgtype.Text{String: req.Email},
		Address: pgtype.Text{String: req.Address},
	})
	if err != nil {
		return nil, err
	}

	return &pb.CustomerResponse{Customer: s.mapCustomerToProto(customer)}, nil
}

// ----- GetCustomer -----
func (s *Service) GetCustomer(ctx context.Context, req *pb.GetCustomerRequest) (*pb.CustomerResponse, error) {
	customer, err := s.queries.GetCustomerByID(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, err
	}
	return &pb.CustomerResponse{Customer: s.mapCustomerToProto(customer)}, nil
}

// ----- ListCustomers -----
func (s *Service) ListCustomers(ctx context.Context, req *pb.ListCustomersRequest) (*pb.ListCustomersResponse, error) {
	customers, err := s.queries.ListCustomers(ctx, db.ListCustomersParams{
		Limit: req.Limit,
		Offset: req.Page * req.Limit,
	})
	if err != nil {
		return nil, err
	}

	pbCustomers := make([]*pb.Customer, 0, len(customers))
	for _, c := range customers {
		pbCustomers = append(pbCustomers, s.mapCustomerToProto(c))
	}

	return &pb.ListCustomersResponse{Customers: pbCustomers}, nil
}

// ----- UpdateCustomer -----
func (s *Service) UpdateCustomer(ctx context.Context, req *pb.UpdateCustomerRequest) (*pb.CustomerResponse, error) {
	customer, err := s.queries.UpdateCustomer(ctx, db.UpdateCustomerParams{
		ID:      req.Id,
		Name:    req.Name,
		Phone:   req.Phone,
		Email:   pgtype.Text{String: req.Email},
		Address: pgtype.Text{String: req.Address},
	})
	if err != nil {
		return nil, err
	}
	return &pb.CustomerResponse{Customer: s.mapCustomerToProto(customer)}, nil
}

// ----- DeleteCustomer -----
func (s *Service) DeleteCustomer(ctx context.Context, req *pb.DeleteCustomerRequest) (*pb.DeleteCustomerResponse, error) {
	err := s.queries.DeleteCustomer(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteCustomerResponse{Success: true}, nil
}
