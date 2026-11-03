package service

import (
	db "github.com/linhhuynhcoding/jss-microservices/product/internal/repository"
	pb "github.com/linhhuynhcoding/jss-microservices/rpc/gen/product"
)

func (s *Service) mapCustomerToProto(c db.Customer) *pb.Customer {
	return &pb.Customer{
		Id:        int32(c.ID),
		Name:      c.Name,
		Phone:     c.Phone,
		Email:     c.Email.String,
		Address:   c.Address.String,
		CreatedAt: c.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: c.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}
