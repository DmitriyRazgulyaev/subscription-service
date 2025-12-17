package logger

import (
	"fmt"
	"go.uber.org/zap"
)

const (
	developmentLevel = "development"
)

type Logger struct {
	z *zap.Logger
}

func NewLogger(environment string) (*Logger, error) {
	loggerCfg := zap.NewProductionConfig()
	loggerCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	if environment == developmentLevel {
		loggerCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := loggerCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("error with building logger: %w", err)
	}
	defer logger.Sync()

	lo := &Logger{z: logger}
	return lo, nil
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.z.Info(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.z.Debug(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.z.Error(msg, fields...)
}
