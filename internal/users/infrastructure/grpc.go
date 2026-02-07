package infrastructure

import (
	"context"

	userspb "go-micro/api/gen/users/v1"
	"go-micro/internal/users/application"
)

// GRPCServer implements the gRPC UserServiceServer
type GRPCServer struct {
	userspb.UnimplementedUserServiceServer
	useCase *application.UserUseCase
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(useCase *application.UserUseCase) *GRPCServer {
	return &GRPCServer{useCase: useCase}
}

// GetUser implements UserServiceServer.GetUser
func (s *GRPCServer) GetUser(ctx context.Context, req *userspb.GetUserRequest) (*userspb.UserResponse, error) {
	output, err := s.useCase.GetUser(ctx, application.GetUserInput{
		ID: uint(req.GetId()),
	})
	if err != nil {
		return nil, err
	}

	return &userspb.UserResponse{
		Id:        uint64(output.User.ID),
		Name:      output.User.Name,
		Email:     output.User.Email,
		CreatedAt: output.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CreateUser implements UserServiceServer.CreateUser
func (s *GRPCServer) CreateUser(ctx context.Context, req *userspb.CreateUserRequest) (*userspb.UserResponse, error) {
	output, err := s.useCase.CreateUser(ctx, application.CreateUserInput{
		Name:  req.GetName(),
		Email: req.GetEmail(),
	})
	if err != nil {
		return nil, err
	}

	return &userspb.UserResponse{
		Id:        uint64(output.User.ID),
		Name:      output.User.Name,
		Email:     output.User.Email,
		CreatedAt: output.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
