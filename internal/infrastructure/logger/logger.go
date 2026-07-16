package logger

import (
	"os"
	"path/filepath"

	"erp/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg config.LogConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.Set(cfg.Level); err != nil {
		level = zapcore.InfoLevel
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	cores := []zapcore.Core{zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)}
	if cfg.Filename != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Filename), 0755); err != nil {
			return nil, err
		}
		file, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(file), level))
	}
	return zap.New(zapcore.NewTee(cores...), zap.AddCaller()), nil
}
