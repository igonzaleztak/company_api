package bootstrap

import (
	"os"
	"xm_test/internal/conf"
	"xm_test/internal/enum"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger creates a new zap logger with the specified log level.
func NewZapLogger() (*zap.SugaredLogger, error) {
	pe := zap.NewProductionEncoderConfig()

	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	level := zap.InfoLevel
	if conf.GlobalConfig.LogLevel == enum.Debug {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	return zap.New(core, zap.AddCaller()).Sugar(), nil
}

var Logger *zap.SugaredLogger
