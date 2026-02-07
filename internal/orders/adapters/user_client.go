package adapters

import (
	"context"

	userspb "go-micro/api/gen/users/v1"
	"go-micro/internal/orders/ports"
	"go-micro/pkg/config"
	grpcpkg "go-micro/pkg/grpc"
	"go-micro/pkg/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCUserClient implements UserClient using gRPC
type GRPCUserClient struct {
	client userspb.UserServiceClient
	conn   *grpc.ClientConn
}

// NewGRPCUserClient creates a new gRPC client for the users service
func NewGRPCUserClient(cfg *config.Config) (*GRPCUserClient, error) {
	var opts []grpc.DialOption

	// Add client interceptor
	opts = append(opts, grpc.WithUnaryInterceptor(grpcpkg.UnaryClientInterceptor(cfg.GRPCTimeout)))

	// Configure TLS/mTLS
	if cfg.GRPCMTLSEnabled {
		tlsConfig, err := tls.ClientConfig(
			"certs/orders-client.crt",
			"certs/orders-client.key",
			cfg.TLSCAFile,
		)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(cfg.UsersGRPCAddr, opts...)
	if err != nil {
		return nil, err
	}

	return &GRPCUserClient{
		client: userspb.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

// GetUser retrieves a user by ID via gRPC
func (c *GRPCUserClient) GetUser(ctx context.Context, userID uint) (*ports.UserInfo, error) {
	resp, err := c.client.GetUser(ctx, &userspb.GetUserRequest{
		Id: uint64(userID),
	})
	if err != nil {
		return nil, err
	}

	return &ports.UserInfo{
		ID:    uint(resp.GetId()),
		Name:  resp.GetName(),
		Email: resp.GetEmail(),
	}, nil
}

// Close closes the gRPC connection
func (c *GRPCUserClient) Close() error {
	return c.conn.Close()
}
