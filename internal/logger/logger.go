package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl
	cfg.EncoderConfig.TimeKey = "timestamp"
    cfg.EncoderConfig.MessageKey = "message"

	configuredLogger, err := cfg.Build()

	if err != nil {
		return err
	}

	Log = configuredLogger
	return nil
}
