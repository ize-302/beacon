package sim

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
	"github.com/ize-302/beacon/backend/internal/vehicles"
)

// poster is what the sender needs from the API. Tests substitute their own.
type poster interface {
	sendGpsPoints(ctx context.Context, points []gpspoints.CreateGpsPoint) error
}

type apiClient struct {
	baseURL string
	http    *http.Client
}

func newAPIClient(baseURL string) *apiClient {
	return &apiClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *apiClient) fetchVehicles() ([]vehicles.VehicleResponse, error) {
	resp, err := c.http.Get(c.baseURL + "/api/v1/vehicles")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch vehicles: unexpected status %s", resp.Status)
	}

	var envelope struct {
		Data []vehicles.VehicleResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// sendGpsPoints posts a whole batch in one request. The slice is marshalled
// before this returns, so the caller is free to reuse its backing array.
func (c *apiClient) sendGpsPoints(ctx context.Context, points []gpspoints.CreateGpsPoint) error {
	if len(points) == 0 {
		return nil
	}

	body, err := json.Marshal(gpspoints.CreateGpsPointsBatch{Points: points})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/gps-points/batch", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("send %d gps points: unexpected status %s", len(points), resp.Status)
	}
	return nil
}

func (c *apiClient) subscribeToNewVehicles(ctx context.Context, onNew func(vehicles.VehicleResponse)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/vehicles/events", nil)
	if err != nil {
		return err
	}
	// The SSE stream is open-ended, so it must not inherit the client timeout.
	resp, err := (&http.Client{}).Do(req)
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
		var v vehicles.VehicleResponse
		if err := json.Unmarshal([]byte(data), &v); err != nil {
			continue
		}
		onNew(v)
	}
	return scanner.Err()
}
