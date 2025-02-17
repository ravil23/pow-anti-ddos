package logx

import (
	"context"
	"os"

	"go.uber.org/zap"
)

var log *zap.Logger

type contextKey string

const (
	ContextKeyRequestID contextKey = "request_id"
	ContextKeyClientNum contextKey = "client_num"
)

func ctxLogger(ctx context.Context) *zap.Logger {
	logger := log
	for _, ctxKey := range []contextKey{ContextKeyRequestID, ContextKeyClientNum} {
		requestID, ok := ctx.Value(ctxKey).(string)
		if ok {
			logger = logger.With(zap.String(string(ctxKey), requestID))
		}
	}
	return logger
}

func Sync() error {
	return log.Sync()
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	ctxLogger(ctx).Info(msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	ctxLogger(ctx).Fatal(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	ctxLogger(ctx).Error(msg, fields...)
}

func init() {
	var err error
	switch os.Getenv("LOG_MODE") {
	case "production":
		log, err = zap.NewProduction()
		if err != nil {
			panic(err)
		}
	default:
		log, err = zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
	}
}
