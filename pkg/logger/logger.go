package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey string

const (
	traceIDKey ctxKey = "trace_id"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	service string
}

// New creates a new logger instance
func New(service, level string) *Logger {
	// Parse log level
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Configure encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	// Create logger with service field
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zapLogger = zapLogger.With(zap.String("service", service))

	return &Logger{
		Logger:  zapLogger,
		service: service,
	}
}

// WithTraceID returns a new logger with the trace ID from context
func (l *Logger) WithTraceID(ctx context.Context) *zap.Logger {
	if traceID := GetTraceID(ctx); traceID != "" {
		return l.With(zap.String("trace_id", traceID))
	}
	return l.Logger
}

// WithContext returns a logger with context fields
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	logger := l.Logger
	if traceID := GetTraceID(ctx); traceID != "" {
		logger = logger.With(zap.String("trace_id", traceID))
	}
	return logger
}

// WithTraceIDContext adds a trace ID to the context
func WithTraceIDContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID retrieves the trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
