// Package api
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ize-302/beacon/backend/cmd/simulator/models"
	"github.com/ize-302/beacon/backend/internal/gps"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
)

func APIFetchAssignments(baseURL string) ([]gps.GpsResponse, error) {
	resp, err := http.Get(baseURL + "/gps")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("status: ", resp.Status)

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	gpss := []gps.GpsResponse{}

	if err = json.Unmarshal(resBody, &gpss); err != nil {
		return nil, err
	}
	return gpss, nil
}

func APISendLocationUpdate(payload models.GpsPayload, baseURL string) {
	tpayload := gpspoints.CreateGpsPoint{GpsID: payload.GpsID, Latitude: payload.Latitude, Longitude: payload.Longitude}
	jsonData, err := json.Marshal(tpayload)
	if err != nil {
		panic(err)
	}
	bodyReader := bytes.NewReader(jsonData)
	_ = bodyReader

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/gps-points", bodyReader)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("GpsID: %d [Lat: %f Lng %f]\n", payload.GpsID, payload.Longitude, payload.Latitude)

	defer resp.Body.Close()
}
