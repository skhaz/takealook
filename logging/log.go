package log

import (
	"go.uber.org/zap"
)

var z *zap.Logger

func init() {
	var err error
	z, err = zap.NewProduction()
	if err != nil {
		panic("failed to initialize zap logger")
	}
}

func Sync() {
	//nolint:golint,errcheck
	z.Sync()
}

func Info(msg string, fields ...zap.Field) {
	z.WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	z.WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	z.WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	z.WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}
