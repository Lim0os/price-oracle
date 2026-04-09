// Package logging provides zap logger initialization and gRPC logging interceptors.
package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a zap logger with the given level.
func NewLogger(level string) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(lvl)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return config.Build()
}

// Field is a generic helper that creates a zap.Field based on the value's type.
func Field[T any](key string, val T) zap.Field {
	switch v := any(val).(type) {
	case string:
		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case int64:
		return zap.Int64(key, v)
	case int32:
		return zap.Int32(key, v)
	case uint:
		return zap.Uint(key, v)
	case uint64:
		return zap.Uint64(key, v)
	case float64:
		return zap.Float64(key, v)
	case float32:
		return zap.Float32(key, v)
	case bool:
		return zap.Bool(key, v)
	case error:
		return zap.Error(v)
	case fmt.Stringer:
		return zap.Stringer(key, v)
	default:
		return zap.Any(key, v)
	}
}

// ZapString creates a zap.String field.
func ZapString(key, val string) zap.Field {
	return zap.String(key, val)
}

// ZapInt creates a zap.Int field.
func ZapInt(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// ZapError creates a zap.Error field.
func ZapError(err error) zap.Field {
	return zap.Error(err)
}
