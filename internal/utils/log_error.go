package utils

import (
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"go.uber.org/zap"
)

func LogErrorWrapper(err error) {
	if err != nil {
		logger.Log.Error("Error wrapped", zap.Error(err))
	}
}
