package infrastructure

import (
	"context"

	orderspb "go-micro/api/gen/orders/v1"
	"go-micro/internal/orders/application"
)

// GRPCServer implements the gRPC OrderServiceServer
type GRPCServer struct {
	orderspb.UnimplementedOrderServiceServer
	useCase *application.OrderUseCase
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(useCase *application.OrderUseCase) *GRPCServer {
	return &GRPCServer{useCase: useCase}
}

// GetOrder implements OrderServiceServer.GetOrder
func (s *GRPCServer) GetOrder(ctx context.Context, req *orderspb.GetOrderRequest) (*orderspb.OrderResponse, error) {
	output, err := s.useCase.GetOrder(ctx, application.GetOrderInput{
		ID: uint(req.GetId()),
	})
	if err != nil {
		return nil, err
	}

	return &orderspb.OrderResponse{
		Id:        uint64(output.Order.ID),
		UserId:    uint64(output.Order.UserID),
		Total:     output.Order.Total,
		Status:    string(output.Order.Status),
		CreatedAt: output.Order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CreateOrder implements OrderServiceServer.CreateOrder
func (s *GRPCServer) CreateOrder(ctx context.Context, req *orderspb.CreateOrderRequest) (*orderspb.OrderResponse, error) {
	output, err := s.useCase.CreateOrder(ctx, application.CreateOrderInput{
		UserID: uint(req.GetUserId()),
		Total:  req.GetTotal(),
	})
	if err != nil {
		return nil, err
	}

	return &orderspb.OrderResponse{
		Id:        uint64(output.Order.ID),
		UserId:    uint64(output.Order.UserID),
		Total:     output.Order.Total,
		Status:    string(output.Order.Status),
		CreatedAt: output.Order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
