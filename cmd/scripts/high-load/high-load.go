package main

import (
	"fmt"
	"sync"

	"github.com/go-resty/resty/v2"
)

func main() {
	client := resty.New()

	g := &sync.WaitGroup{}

	for range 10 {
		g.Add(1)
		go func() {
			for range 10_000 {
				_, err := client.R().SetBody(`{"url": "https://google.com"}`).Post("http://localhost:8080/api/shorten/")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			g.Done()
			}()
		}
	g.Wait()
}
