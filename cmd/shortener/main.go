package main

import (
	"fmt"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/handler"
)

func main() {
	handler.Router.Run(fmt.Sprintf(":%d", config.Config.Port))
}
