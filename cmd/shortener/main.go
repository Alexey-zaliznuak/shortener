package main

import (
	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/handler"
)

func main() {
	handler.Router.Run(config.Config.StartupAddress)
}
