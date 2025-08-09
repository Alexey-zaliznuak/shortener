package main

import (
	"flag"
	"fmt"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/handler"
	"github.com/Alexey-zaliznuak/shortener/internal/logger"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
	"go.uber.org/zap"
)

func main() {
	flagsConfig := config.CreateFLagsInitialConfig()
	flag.Parse()

	cfg, err := config.GetConfig(flagsConfig)

	if err != nil {
		fmt.Println(err.Error())
	}

	logger.Initialize(cfg.LoggingLevel)

	logger.Log.Info("Configuration", zap.Any("config", cfg))

	linksService := &service.LinksService{AppConfig: cfg}

	router := handler.NewRouter()
	handler.RegisterLinksRoutes(router, linksService)

	router.Run(cfg.ServerAddress)
}
