package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"github.com/ize-302/beacon/backend/cmd/simulator/api"
	"github.com/ize-302/beacon/backend/cmd/simulator/models"
	"github.com/joho/godotenv"
)

var files = []string{"abayomi-drive.json", "adekunle-street-yaba.json", "adetola-street-yaba.json", "akoka-road.json", "commercial-avenue-yaba.json", "iwaya-road-yaba.json", "sabo-road-yaba.json", "tejuosho-street.json", "university-road.json", "yaba-herbert-macaulay-way.json"}

var baseURL string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	baseURL = fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))

	var wg sync.WaitGroup
	var mu sync.Mutex

	assignments, err := api.APIFetchAssignments(baseURL)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	jobsChan := make(chan string, len(files))
	routesChan := make(chan [][]float64, len(files))
	gpsChan := make(chan models.Gps, len(assignments))

	// queue files as jobs in channel
	for _, file := range files {
		jobsChan <- file
	}
	close(jobsChan)

	// spinup 3 go routines to process routes from jobs
	for i := range 3 {
		wg.Add(1)
		go fileReaderWorker(i, jobsChan, routesChan, &mu, &wg)
	}

	go func() {
		wg.Wait()
		close(routesChan)
	}()

	// populate each vehicle with routes
	for _, a := range assignments {
		gpsChan <- models.Gps{
			AssignmentID: a.ID,
			Routes:       <-routesChan,
		}
	}
	close(gpsChan)

	// spinup len(vehicles) goroutine to manage tick for each gps
	var tickerWg sync.WaitGroup
	for i := range len(assignments) {
		tickerWg.Add(1)
		go ticker(i, gpsChan, &tickerWg)
	}
	tickerWg.Wait()
}

func fileReaderWorker(w int, jobsChan chan string, routesChan chan [][]float64, mu *sync.Mutex, wg *sync.WaitGroup) {
	_ = w
	defer wg.Done()
	for job := range jobsChan {
		path := fmt.Sprintf("cmd/simulator/data/%v", job)
		_ = mu
		mu.Lock()
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		var route models.Route
		err = json.Unmarshal(content, &route)
		if err != nil {
			log.Fatal(err)
		}
		mu.Unlock()
		// message := fmt.Sprintf("wkr %d => %s route\n", w, route.Name)
		// fmt.Println(message)
		routesChan <- route.Coordinates
	}
}

func ticker(w int, gpsChan chan models.Gps, wg *sync.WaitGroup) {
	_ = w
	defer wg.Done()
	for gps := range gpsChan {
		randomSpeed := rand.IntN(9) + 1

		t := time.NewTicker(time.Duration(randomSpeed) * time.Second)

		for range t.C {
			gps.CurrentIndex++

			if gps.CurrentIndex >= len(gps.Routes) {
				gps.CurrentIndex = 0
			}

			location := gps.Routes[gps.CurrentIndex]
			gpsPayload := models.GpsPayload{
				AssignmentID: gps.AssignmentID,
				Latitude:     location[0],
				Longitude:    location[1],
			}

			api.APISendLocationUpdate(gpsPayload, baseURL)
		}
		t.Stop()
	}
}
