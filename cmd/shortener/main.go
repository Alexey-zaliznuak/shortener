package main

import (
	"flag"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/handler"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
)

func main() {
	flagsConfig := config.CreateFLagsInitialConfig()
	flag.Parse()

	cfg := config.GetConfig(flagsConfig)

	linksService := &service.LinksService{AppConfig: cfg}

	router := handler.NewRouter()
	handler.RegisterLinksRoutes(router, linksService)

	router.Run(cfg.StartupAddress)
}
