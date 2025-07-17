package main

import (
	"flag"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/handler"
)

func main() {
	flag.Parse()
	handler.Router.Run(config.GetConfig().StartupAddress)
}
