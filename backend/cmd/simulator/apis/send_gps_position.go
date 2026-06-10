// Package apis
package apis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
)

func APISendGpsPosition(payload gpspoints.CreateGpsPoint, baseURL string) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	bodyReader := bytes.NewReader(jsonData)

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
