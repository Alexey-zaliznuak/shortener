package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

func main() {
	goroutines := 10
	total := 1_000

	client := resty.New()

	g := &sync.WaitGroup{}

	start := time.Now()

	for range goroutines {
		g.Add(1)
		go func() {
			for range total / goroutines {
				_, err := client.R().SetBody(`{"url": "https://google.com"}`).Post("http://localhost:8080/api/shorten/")
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			g.Done()
		}()
	}
	g.Wait()
	end := time.Now()
	fmt.Printf("Average RPS: %f\n", float64(total)/float64(end.Sub(start).Seconds()))
}
