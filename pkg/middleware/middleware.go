package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-micro/pkg/errors"
	"go-micro/pkg/logger"
)

const (
	// TraceIDHeader is the header name for trace ID
	TraceIDHeader = "X-Trace-ID"
	// TraceIDKey is the context key for trace ID
	TraceIDKey = "trace_id"
)

// ErrorHandler is a middleware that handles errors and panics
func ErrorHandler(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				traceID := c.GetString(TraceIDKey)
				log.WithContext(c.Request.Context()).Error("panic recovered",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
					zap.String("trace_id", traceID),
				)

				c.Header(TraceIDHeader, traceID)
				c.AbortWithStatusJSON(http.StatusInternalServerError, errors.ErrorResponse{
					Error: errors.ErrorBody{
						Code:    errors.CodeInternal,
						Message: "An internal error occurred",
					},
					TraceID: traceID,
				})
			}
		}()

		c.Next()

		// Handle errors set by handlers
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			traceID := c.GetString(TraceIDKey)
			statusCode, jsonResponse := errors.ToJSON(err, traceID)

			log.WithContext(c.Request.Context()).Error("request error",
				zap.Error(err),
				zap.Int("status", statusCode),
				zap.String("trace_id", traceID),
			)

			c.Header(TraceIDHeader, traceID)
			c.Data(statusCode, "application/json", jsonResponse)
		}
	}
}

// TraceID is a middleware that generates or extracts trace ID
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDHeader, traceID)

		// Add trace ID to request context
		ctx := logger.WithTraceIDContext(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequestLogger logs all HTTP requests
func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		traceID := c.GetString(TraceIDKey)

		log.WithContext(c.Request.Context()).Info("http request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("trace_id", traceID),
		)
	}
}

// CORS is a middleware that handles CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Trace-ID")
		c.Header("Access-Control-Expose-Headers", "X-Trace-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
