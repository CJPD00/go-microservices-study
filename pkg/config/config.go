package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	ServiceName string

	// HTTP
	HTTPPort  string
	HTTPSPort string

	// gRPC
	GRPCPort       string
	UsersGRPCAddr  string
	OrdersGRPCAddr string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// RabbitMQ
	RabbitMQURL string

	// TLS
	TLSEnabled      bool
	TLSCertFile     string
	TLSKeyFile      string
	TLSCAFile       string
	GRPCMTLSEnabled bool
	GRPCClientCert  string
	GRPCClientKey   string

	// Logging
	LogLevel  string
	LogFormat string

	// Timeouts
	DBTimeout   time.Duration
	GRPCTimeout time.Duration
	HTTPTimeout time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if exists (ignore error if not found)
	_ = godotenv.Load()

	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "service"),

		// HTTP
		HTTPPort:  getEnv("HTTP_PORT", "8080"),
		HTTPSPort: getEnv("HTTPS_PORT", "8443"),

		// gRPC
		GRPCPort:       getEnv("GRPC_PORT", "50051"),
		UsersGRPCAddr:  getEnv("USERS_GRPC_ADDR", "localhost:50051"),
		OrdersGRPCAddr: getEnv("ORDERS_GRPC_ADDR", "localhost:50052"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "postgres"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// RabbitMQ
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),

		// TLS
		TLSEnabled:      getEnvBool("TLS_ENABLED", false),
		TLSCertFile:     getEnv("TLS_CERT_FILE", "certs/gateway.crt"),
		TLSKeyFile:      getEnv("TLS_KEY_FILE", "certs/gateway.key"),
		TLSCAFile:       getEnv("TLS_CA_FILE", "certs/ca.crt"),
		GRPCMTLSEnabled: getEnvBool("GRPC_MTLS_ENABLED", false),
		GRPCClientCert:  getEnv("GRPC_CLIENT_CERT_FILE", "certs/gateway-client.crt"),
		GRPCClientKey:   getEnv("GRPC_CLIENT_KEY_FILE", "certs/gateway-client.key"),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		// Timeouts
		DBTimeout:   getEnvDuration("DB_TIMEOUT", 30*time.Second),
		GRPCTimeout: getEnvDuration("GRPC_TIMEOUT", 10*time.Second),
		HTTPTimeout: getEnvDuration("HTTP_TIMEOUT", 30*time.Second),
	}
}

// LoadForService loads configuration with service-specific overrides
func LoadForService(serviceName string) *Config {
	_ = godotenv.Load()

	cfg := Load()
	cfg.ServiceName = serviceName

	// Override database config based on service
	prefix := serviceName + "_"
	if v := os.Getenv(prefix + "DB_HOST"); v != "" {
		cfg.DBHost = v
	}
	if v := os.Getenv(prefix + "DB_PORT"); v != "" {
		cfg.DBPort = v
	}
	if v := os.Getenv(prefix + "DB_USER"); v != "" {
		cfg.DBUser = v
	}
	if v := os.Getenv(prefix + "DB_PASSWORD"); v != "" {
		cfg.DBPassword = v
	}
	if v := os.Getenv(prefix + "DB_NAME"); v != "" {
		cfg.DBName = v
	}

	return cfg
}

// DSN returns the database connection string
func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		seconds, err := strconv.Atoi(value)
		if err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}
