// Package main Go-Micro Gateway API
//
// This is the API Gateway for Go-Micro microservices project.
// It provides REST API endpoints that communicate with internal gRPC services.
//
//	@title			Go-Micro Gateway API
//	@version		1.0
//	@description	API Gateway for Go-Micro microservices
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.email	support@example.com
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8443
//	@BasePath	/
//	@schemes	https http
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-micro/docs/swagger"
	"go-micro/internal/gateway/clients"
	"go-micro/internal/gateway/handlers"
	"go-micro/pkg/config"
	"go-micro/pkg/logger"
	"go-micro/pkg/middleware"
	pkgtls "go-micro/pkg/tls"
)

func main() {
	// Load configuration
	cfg := config.Load()
	cfg.ServiceName = "gateway"

	// Initialize logger
	log := logger.New("gateway", cfg.LogLevel)
	defer log.Sync()

	log.Info("starting gateway service")

	// Create gRPC clients
	grpcClients, err := clients.NewClients(cfg)
	if err != nil {
		log.Fatal("failed to create gRPC clients: " + err.Error())
	}
	defer grpcClients.Close()
	log.Info("connected to backend services via gRPC")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.TraceID())
	router.Use(middleware.RequestLogger(log))
	router.Use(middleware.ErrorHandler(log))
	router.Use(middleware.CORS())

	// Register API routes
	handler := handlers.NewHandler(grpcClients.Users, grpcClients.Orders)
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Root redirect to Swagger
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})

	// Start server
	if cfg.TLSEnabled {
		startHTTPSServer(cfg, log, router, ctx)
	} else {
		startHTTPServer(cfg, log, router, ctx)
	}
}

func startHTTPServer(cfg *config.Config, log *logger.Logger, router *gin.Engine, ctx context.Context) {
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  cfg.HTTPTimeout,
		WriteTimeout: cfg.HTTPTimeout,
	}

	go func() {
		log.Info("HTTP server listening on http://localhost:" + cfg.HTTPPort)
		log.Info("Swagger UI: http://localhost:" + cfg.HTTPPort + "/swagger/index.html")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error: " + err.Error())
		}
	}()

	waitForShutdown(server, log, ctx)
}

func startHTTPSServer(cfg *config.Config, log *logger.Logger, router *gin.Engine, ctx context.Context) {
	tlsConfig, err := pkgtls.ServerConfig(cfg.TLSCertFile, cfg.TLSKeyFile, "", false)
	if err != nil {
		log.Fatal("failed to load TLS config: " + err.Error())
	}

	server := &http.Server{
		Addr:         ":" + cfg.HTTPSPort,
		Handler:      router,
		TLSConfig:    tlsConfig,
		ReadTimeout:  cfg.HTTPTimeout,
		WriteTimeout: cfg.HTTPTimeout,
	}

	go func() {
		log.Info("HTTPS server listening on https://localhost:" + cfg.HTTPSPort)
		log.Info("Swagger UI: https://localhost:" + cfg.HTTPSPort + "/swagger/index.html")
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTPS server error: " + err.Error())
		}
	}()

	waitForShutdown(server, log, ctx)
}

func waitForShutdown(server *http.Server, log *logger.Logger, ctx context.Context) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error: " + err.Error())
	}

	log.Info("server stopped")
}

// Ensure tls.Config is used to avoid unused import
var _ *tls.Config
