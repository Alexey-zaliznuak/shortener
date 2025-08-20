package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type CreateBatchRequest struct {
	URL           string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

// При тесте на 1000 000 запросах/размере батча
// Судя по всему без распараллеливания в несколько потоков батч менее эффективен
// батч - вставка в секунду: postgres=598 in_memory=47816
// одиночные запросы - вставка в секунду: postgres=1159 in_memory=7000

func createSingleLink(total int) {
	goroutines := 10

	if total == 0 {
		return
	}

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

func createBigBatch(total int) {
	if total == 0 {
		return
	}

	client := resty.New()

	var data []CreateBatchRequest

	for i := range total {
		data = append(data, CreateBatchRequest{URL: "https://ya.ru", CorrelationID: strconv.Itoa(i)})
	}

	var body bytes.Buffer

	err := json.NewEncoder(&body).Encode(data)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	start := time.Now()
	_, err = client.R().SetBody(body.Bytes()).Post("http://localhost:8080/api/shorten/batch")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	end := time.Now()
	fmt.Printf("Batch: average rows write per second: %f\n", float64(total)/float64(end.Sub(start).Seconds()))
}

func main() {
	createSingleLinkRequests := 0
	createBatchRows := 1000

	g := &sync.WaitGroup{}

	g.Add(1)
	go func() {
		createSingleLink(createSingleLinkRequests)
		g.Done()
	}()

	g.Add(1)
	go func() {
		createBigBatch(createBatchRows)
		g.Done()
	}()

	g.Wait()
}
