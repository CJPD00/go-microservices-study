package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	userspb "go-micro/api/gen/users/v1"
	"go-micro/internal/users/adapters"
	"go-micro/internal/users/application"
	"go-micro/internal/users/infrastructure"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/events"
	grpcpkg "go-micro/pkg/grpc"
	"go-micro/pkg/logger"
	"go-micro/pkg/middleware"
	"go-micro/pkg/rabbitmq"
	"go-micro/pkg/tls"
)

func main() {
	// Load configuration
	cfg := config.LoadForService("USERS")
	cfg.DBHost = getEnvOrDefault("USERS_DB_HOST", "localhost")
	cfg.DBPort = getEnvOrDefault("USERS_DB_PORT", "5432")
	cfg.DBName = getEnvOrDefault("USERS_DB_NAME", "users_db")
	cfg.GRPCPort = getEnvOrDefault("USERS_GRPC_PORT", "50051")

	// Initialize logger
	log := logger.New("users-service", cfg.LogLevel)
	defer log.Sync()

	log.Info("starting users service")

	// Connect to database
	dbConn, err := db.NewConnection(db.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
		Timeout:  cfg.DBTimeout,
	})
	if err != nil {
		log.Fatal("failed to connect to database: " + err.Error())
	}
	log.Info("connected to database")

	// Initialize repository and run migrations
	repo := adapters.NewPostgresUserRepository(dbConn)
	if err := repo.Migrate(); err != nil {
		log.Fatal("failed to migrate database: " + err.Error())
	}

	// Connect to RabbitMQ
	var publisher *adapters.RabbitMQPublisher
	rabbitConn, err := rabbitmq.NewConnection(cfg.RabbitMQURL, log)
	if err != nil {
		log.Warn("failed to connect to RabbitMQ, events will be disabled: " + err.Error())
	} else {
		defer rabbitConn.Close()
		pub, err := rabbitmq.NewPublisher(rabbitConn, events.ExchangeUsers, log)
		if err != nil {
			log.Warn("failed to create publisher: " + err.Error())
		} else {
			publisher = adapters.NewRabbitMQPublisher(pub, log)
		}
	}

	// Initialize use case
	useCase := application.NewUserUseCase(repo, publisher, log)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server
	httpHandler := infrastructure.NewHTTPHandler(useCase)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.TraceID())
	router.Use(middleware.RequestLogger(log))
	router.Use(middleware.ErrorHandler(log))
	router.Use(middleware.CORS())

	api := router.Group("/api/v1")
	httpHandler.RegisterRoutes(api)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  cfg.HTTPTimeout,
		WriteTimeout: cfg.HTTPTimeout,
	}

	go func() {
		log.Info("HTTP server listening on :" + cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error: " + err.Error())
		}
	}()

	// Start gRPC server
	grpcServer := setupGRPCServer(cfg, log, useCase)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen for gRPC: " + err.Error())
	}

	go func() {
		log.Info("gRPC server listening on :" + cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("gRPC server error: " + err.Error())
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down servers...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP shutdown error: " + err.Error())
	}

	log.Info("servers stopped")
}

func setupGRPCServer(cfg *config.Config, log *logger.Logger, useCase *application.UserUseCase) *grpc.Server {
	var opts []grpc.ServerOption

	// Add interceptors
	opts = append(opts, grpc.UnaryInterceptor(grpcpkg.UnaryServerInterceptor(log, cfg.GRPCTimeout)))

	// Configure mTLS if enabled
	if cfg.GRPCMTLSEnabled {
		tlsConfig, err := tls.ServerConfig(
			"certs/users.crt",
			"certs/users.key",
			cfg.TLSCAFile,
			true, // require client cert
		)
		if err != nil {
			log.Fatal("failed to load TLS config: " + err.Error())
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		log.Info("gRPC mTLS enabled")
	}

	server := grpc.NewServer(opts...)
	userspb.RegisterUserServiceServer(server, infrastructure.NewGRPCServer(useCase))

	return server
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
