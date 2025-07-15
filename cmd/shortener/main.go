package main

import (
	"github.com/Alexey-zaliznuak/shortener/internal/handler"
)

func main() {
	handler.Router.Run(":8080")
}
