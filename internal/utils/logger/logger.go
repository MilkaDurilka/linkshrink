package logger

import "go.uber.org/zap"

type Logger interface {
	With(fields ...zap.Field) *zap.Logger
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}
