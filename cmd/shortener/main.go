package main

import (
	"net/http"

	"github.com/Alexey-zaliznuak/shortener/internal/handler"
)

func main() {
	err := http.ListenAndServe(`:8080`, handler.Mux)
	if err != nil {
		panic(err)
	}
}
