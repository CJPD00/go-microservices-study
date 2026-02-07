package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go-micro/pkg/errors"
	"go-micro/pkg/logger"
)

const (
	// TraceIDMetadataKey is the metadata key for trace ID
	TraceIDMetadataKey = "x-trace-id"
)

// UnaryServerInterceptor creates a server interceptor for logging, tracing, and error handling
func UnaryServerInterceptor(log *logger.Logger, timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Extract or generate trace ID
		traceID := extractTraceID(ctx)
		if traceID == "" {
			traceID = uuid.New().String()
		}
		ctx = logger.WithTraceIDContext(ctx, traceID)

		// Apply timeout
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		// Call handler
		resp, err := handler(ctx, req)

		// Log request
		duration := time.Since(start)
		logFields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("trace_id", traceID),
		}

		if err != nil {
			st, _ := status.FromError(err)
			logFields = append(logFields, zap.String("grpc_code", st.Code().String()))
			log.WithContext(ctx).Error("grpc request failed", logFields...)

			// Convert domain errors to gRPC status
			return nil, errors.GRPCStatus(err)
		}

		log.WithContext(ctx).Info("grpc request completed", logFields...)
		return resp, nil
	}
}

// UnaryClientInterceptor creates a client interceptor for tracing and timeout
func UnaryClientInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Propagate trace ID
		traceID := logger.GetTraceID(ctx)
		if traceID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, TraceIDMetadataKey, traceID)
		}

		// Apply timeout
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			// Convert gRPC status to domain error
			return errors.FromGRPCStatus(err)
		}

		return nil
	}
}

// StreamServerInterceptor creates a stream server interceptor
func StreamServerInterceptor(log *logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		ctx := ss.Context()

		// Extract trace ID
		traceID := extractTraceID(ctx)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		err := handler(srv, ss)

		duration := time.Since(start)
		log.WithContext(ctx).Info("grpc stream completed",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("trace_id", traceID),
			zap.Error(err),
		)

		return err
	}
}

func extractTraceID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(TraceIDMetadataKey)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
