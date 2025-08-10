package main

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

func main() {
	client := resty.New()

	for range 100_000 {
		_, err := client.R().SetBody(`{"url": "https://google.com"}`).Post("http://localhost:8080/api/shorten")
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
