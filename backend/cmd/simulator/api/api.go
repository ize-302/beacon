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
	"github.com/ize-302/beacon/backend/internal/assignments"
	"github.com/ize-302/beacon/backend/internal/locations"
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

func APIFetchAssignments() ([]assignments.AssignmentResponse, error) {
	baseURL := handleBaseURL()
	resp, err := http.Get(baseURL + "/vehicle-assignments")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("status: ", resp.Status)

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	assignments := []assignments.AssignmentResponse{}

	if err = json.Unmarshal(resBody, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func APISendLocationUpdate(payload models.GpsPayload) {
	tpayload := locations.CreateLocation{AssignmentID: payload.AssignmentID, Latitude: payload.Latitude, Longitude: payload.Longitude}
	jsonData, err := json.Marshal(tpayload)
	if err != nil {
		panic(err)
	}
	bodyReader := bytes.NewReader(jsonData)
	_ = bodyReader

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

	fmt.Printf("Assignment: %d [Lat: %f Lng %f]\n", payload.AssignmentID, payload.Longitude, payload.Latitude)

	defer resp.Body.Close()
}
