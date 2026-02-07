package clients

import (
	"go-micro/pkg/config"
	grpcpkg "go-micro/pkg/grpc"
	"go-micro/pkg/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	orderspb "go-micro/api/gen/orders/v1"
	userspb "go-micro/api/gen/users/v1"
)

// Clients holds all gRPC clients for the gateway
type Clients struct {
	Users  userspb.UserServiceClient
	Orders orderspb.OrderServiceClient

	usersConn  *grpc.ClientConn
	ordersConn *grpc.ClientConn
}

// NewClients creates all gRPC clients for the gateway
func NewClients(cfg *config.Config) (*Clients, error) {
	// Create users client
	usersConn, err := createConnection(cfg, cfg.UsersGRPCAddr)
	if err != nil {
		return nil, err
	}

	// Create orders client
	ordersConn, err := createConnection(cfg, cfg.OrdersGRPCAddr)
	if err != nil {
		usersConn.Close()
		return nil, err
	}

	return &Clients{
		Users:      userspb.NewUserServiceClient(usersConn),
		Orders:     orderspb.NewOrderServiceClient(ordersConn),
		usersConn:  usersConn,
		ordersConn: ordersConn,
	}, nil
}

// Close closes all gRPC connections
func (c *Clients) Close() error {
	if c.usersConn != nil {
		c.usersConn.Close()
	}
	if c.ordersConn != nil {
		c.ordersConn.Close()
	}
	return nil
}

func createConnection(cfg *config.Config, addr string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	// Add client interceptor
	opts = append(opts, grpc.WithUnaryInterceptor(grpcpkg.UnaryClientInterceptor(cfg.GRPCTimeout)))

	// Configure TLS/mTLS
	if cfg.GRPCMTLSEnabled {
		tlsConfig, err := tls.ClientConfig(
			cfg.GRPCClientCert,
			cfg.GRPCClientKey,
			cfg.TLSCAFile,
		)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return grpc.Dial(addr, opts...)
}
