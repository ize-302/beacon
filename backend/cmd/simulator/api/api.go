// Package api
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ize-302/beacon/backend/cmd/simulator/models"
	"github.com/ize-302/beacon/backend/internal/locations"
	"github.com/ize-302/beacon/backend/internal/vehicles"
	"github.com/joho/godotenv"
)

func handleBaseURL() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	return fmt.Sprintf("http://localhost:%s", port)
}

func APIFetchVehicles() ([]vehicles.VehicleResponse, error) {
	baseURL := handleBaseURL()
	resp, err := http.Get(baseURL + "/vehicles")
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

func APISendLocationUpdate(payload models.GpsPayload) {
	fmt.Printf("Vehicle: %d [Lat: %f Lng %f]\n", payload.VehicleID, payload.Longitude, payload.Latitude)
	tpayload := locations.CreateLocation{VehicleID: payload.VehicleID, Latitude: payload.Latitude, Longitude: payload.Longitude}
	jsonData, err := json.Marshal(tpayload)
	if err != nil {
		panic(err)
	}
	bodyReader := bytes.NewReader(jsonData)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	baseURL := handleBaseURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/locations", bodyReader)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
