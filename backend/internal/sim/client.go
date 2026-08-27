package sim

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	internalgps "github.com/ize-302/beacon/backend/internal/gps"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
)

func fetchGpsDevices(baseURL string) ([]internalgps.GpsResponse, error) {
	resp, err := http.Get(baseURL + "/api/v1/gps-devices")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("status: ", resp.Status)

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data []internalgps.GpsResponse `json:"data"`
	}
	if err = json.Unmarshal(resBody, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

func sendGpsPosition(payload gpspoints.CreateGpsPoint, baseURL string) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	bodyReader := bytes.NewReader(jsonData)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/gps-points", bodyReader)
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

func subscribeToNewDevices(ctx context.Context, baseURL string, onNew func(internalgps.GpsResponse)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/gps-devices/events", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var gps internalgps.GpsResponse
		if err := json.Unmarshal([]byte(data), &gps); err != nil {
			continue
		}
		onNew(gps)
	}
	return scanner.Err()
}
