package main

import (
	"flag"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/handler"
	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/Alexey-zaliznuak/shortener/internal/service"
)

func main() {
	flag.Parse()

	db := database.GetClient()

	linksService := service.NewLinksService(db)

	handler.SetupLinksRoutes(linksService)

	handler.Router.Run(config.GetConfig().StartupAddress)
}
