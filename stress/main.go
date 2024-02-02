package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL  = "http://localhost:8080"
	workers  = 10 // Number of concurrent workers
	requests = 100 // Number of requests per worker
)

func main() {
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go makeRequests(&wg)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("All requests completed in %v\n", elapsed)
}

func makeRequests(wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{}

	for i := 0; i < requests; i++ {
		// Send GET request to /coffees endpoint
		resp, err := client.Get(fmt.Sprintf("%s/coffees", baseURL))
		if err != nil {
			log.Printf("Error making GET request: %v\n", err)
			continue
		}
		resp.Body.Close()

		// Send POST request to /buy endpoint with a random coffee ID
		resp, err = client.Post(fmt.Sprintf("%s/buy", baseURL), "application/json", nil)
		if err != nil {
			log.Printf("Error making POST request: %v\n", err)
			continue
		}
		resp.Body.Close()

		log.Printf("Request %d completed\n", i+1)
	}
}
