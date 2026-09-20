// Command seed creates a fixed number of vehicles through the running API.
//
// It goes through POST /vehicles rather than writing to Postgres so the API
// announces each new vehicle and a running simulator picks it up straight away.
//
//	go run ./cmd/seed -n 50
//	go run ./cmd/seed -n 50 -url https://your-app.up.railway.app
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultURL  = "http://127.0.0.1:8081"
	maxAttempts = 3 // a fresh plate is drawn per attempt, so this only covers rare collisions
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	count := flag.Int("n", 10, "number of vehicles to create")
	baseURL := flag.String("url", envOr("API_BASE_URL", defaultURL), "API base URL")
	workers := flag.Int("c", 8, "concurrent requests")
	flag.Parse()

	if *count < 1 || *workers < 1 {
		log.Fatal("-n and -c must be at least 1")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	endpoint := strings.TrimRight(*baseURL, "/") + "/api/v1/vehicles"

	jobs := make(chan struct{})
	var created, failed atomic.Int64
	var wg sync.WaitGroup

	for range min(*workers, *count) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				if err := createVehicle(client, endpoint); err != nil {
					log.Printf("seed: %v", err)
					failed.Add(1)
					continue
				}
				created.Add(1)
			}
		}()
	}

	for range *count {
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Wait()

	fmt.Printf("created %d/%d vehicles against %s\n", created.Load(), *count, *baseURL)
	if failed.Load() > 0 {
		os.Exit(1)
	}
}

func createVehicle(client *http.Client, endpoint string) error {
	var lastErr error
	for range maxAttempts {
		body, err := json.Marshal(map[string]string{"plate_number": randomPlate()})
		if err != nil {
			return err
		}

		resp, err := client.Post(endpoint, "application/json", bytes.NewReader(body))
		if err != nil {
			return err
		}
		msg, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			return nil
		}
		lastErr = fmt.Errorf("create vehicle: %s: %s", resp.Status, strings.TrimSpace(string(msg)))
	}
	return lastErr
}

func randomPlate() string {
	letters := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte('A' + rand.IntN(26))
		}
		return string(b)
	}
	return fmt.Sprintf("%s-%03d-%s", letters(3), rand.IntN(1000), letters(2))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
