package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ize-302/beacon/backend/cmd/simulator/models"
	"github.com/ize-302/beacon/backend/internal/vehicles"
)

var files = []string{"abayomi-drive.json", "adekunle-street-yaba.json", "adetola-street-yaba.json", "akoka-road.json", "commercial-avenue-yaba.json", "iwaya-road-yaba.json", "sabo-road-yaba.json", "tejuosho-street.json", "university-road.json", "yaba-herbert-macaulay-way.json"}

func fileReaderWorker(w int, jobsChan chan string, routesChan chan [][]float64, mu *sync.Mutex, wg *sync.WaitGroup) {
	_ = w
	defer wg.Done()
	for job := range jobsChan {
		path := fmt.Sprintf("./data/%v", job)
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

func fetchVehicles() ([]vehicles.VehicleResponse, error) {
	resp, err := http.Get("http://localhost:8080/vehicles")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("status: ", resp.Status)

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	vehicles := []vehicles.VehicleResponse{}

	if err = json.Unmarshal(resBody, &vehicles); err != nil {
		return nil, err
	}
	return vehicles, nil
}

func sendLocationUpdate(payload models.GpsPayload) {
	fmt.Printf("Vehicle: %d [Lat: %f Lng %f]\n", payload.VehicleID, payload.Latitude, payload.Longitude)
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
				VehicleID: gps.VehicleID,
				Latitude:  location[0],
				Longitude: location[1],
			}

			sendLocationUpdate(gpsPayload)
		}
		t.Stop()
	}
}

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex

	vehicles, err := fetchVehicles()
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	jobsChan := make(chan string, len(files))
	routesChan := make(chan [][]float64, len(files))
	gpsChan := make(chan models.Gps, len(vehicles))

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
	for _, v := range vehicles {
		gpsChan <- models.Gps{
			VehicleID: v.ID,
			Routes:    <-routesChan,
		}
	}
	close(gpsChan)

	// spinup len(vehicles) goroutine to manage tick for each gps
	var tickerWg sync.WaitGroup
	for i := range len(vehicles) {
		tickerWg.Add(1)
		go ticker(i, gpsChan, &tickerWg)
	}
	tickerWg.Wait()
}
