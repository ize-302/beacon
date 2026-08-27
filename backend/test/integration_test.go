//go:build integration

// Package integration holds black-box tests that drive a running Beacon stack
// over HTTP and WebSocket. They are excluded from the normal build by the
// `integration` tag, since they need the API, database and simulator up:
//
//	docker compose up -d
//	go test -tags=integration ./test/...
//
// Override the targets with API_BASE_URL and WS_URL if the stack is elsewhere.
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type point struct {
	GpsID     int     `json:"gps_id"`
	Bearing   float64 `json:"bearing"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func apiURL() string { return env("API_BASE_URL", "http://localhost:8081") }
func wsURL() string  { return env("WS_URL", "ws://localhost:8081/ws") }

// requireStack skips the test rather than failing it when nothing is running,
// so `go test -tags=integration ./...` on a dev machine without the stack up is
// a no-op instead of a wall of red.
func requireStack(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(apiURL() + "/api/v1/health")
	if err != nil {
		t.Skipf("stack not reachable at %s (%v) — start it with `docker compose up -d`", apiURL(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("health check returned %s", resp.Status)
	}
}

// existingGpsID reuses a registered device rather than creating one. Points
// carry a foreign key to gps_devices, and DELETE /gps-devices is currently a
// no-op (see docs/rearchitecture.md §2), so creating our own would leave
// permanent debris and hand the simulator an extra vehicle to drive.
func existingGpsID(t *testing.T) int {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL() + "/api/v1/gps-devices")
	if err != nil {
		t.Fatalf("list gps devices: %v", err)
	}
	defer resp.Body.Close()

	var envelope struct {
		Data []struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode gps devices: %v", err)
	}
	if len(envelope.Data) == 0 {
		t.Skip("no gps devices registered; create one in the dashboard first")
	}
	return envelope.Data[0].ID
}

func postPoint(t *testing.T, p point) {
	t.Helper()
	body, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(apiURL()+"/api/v1/gps-points", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post point: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("post point returned %s, want 201", resp.Status)
	}
}

func dial(t *testing.T, name string) *websocket.Conn {
	t.Helper()
	c, _, err := websocket.DefaultDialer.Dial(wsURL(), nil)
	if err != nil {
		t.Fatalf("%s could not connect: %v", name, err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

// awaitPoints reads until it has seen every timestamp in want, ignoring the
// simulator's own traffic. Returns them in arrival order.
func awaitPoints(t *testing.T, c *websocket.Conn, want map[int64]bool, timeout time.Duration) []point {
	t.Helper()
	if err := c.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		t.Fatal(err)
	}

	var got []point
	remaining := len(want)
	for remaining > 0 {
		var p point
		if err := c.ReadJSON(&p); err != nil {
			t.Fatalf("expected %d more of our points, read failed: %v", remaining, err)
		}
		if want[p.Timestamp] {
			got = append(got, p)
			remaining--
		}
	}
	return got
}

func TestStackIsUp(t *testing.T) {
	requireStack(t)
}

// A point posted over HTTP must reach connected WebSocket clients.
func TestPointReachesWebSocketClient(t *testing.T) {
	requireStack(t)
	gpsID := existingGpsID(t)

	c := dial(t, "client")
	time.Sleep(200 * time.Millisecond) // let the hub register the connection

	// A distinctive timestamp separates our points from the simulator's.
	base := time.Now().UnixMilli()
	want := map[int64]bool{base: true}
	postPoint(t, point{GpsID: gpsID, Latitude: 6.5, Longitude: 3.3, Bearing: 90, Timestamp: base})

	got := awaitPoints(t, c, want, 10*time.Second)
	if len(got) != 1 {
		t.Fatalf("got %d points, want 1", len(got))
	}
	if got[0].GpsID != gpsID {
		t.Fatalf("got GpsID %d, want %d", got[0].GpsID, gpsID)
	}
}

// The regression test for the hub deadlock: one client vanishing must not stop
// delivery to the others, reorder anything, or stop new clients connecting.
func TestHubSurvivesClientDisconnect(t *testing.T) {
	requireStack(t)
	gpsID := existingGpsID(t)

	dead := dial(t, "client-A")
	live := dial(t, "client-B")
	time.Sleep(200 * time.Millisecond)

	dead.Close()
	time.Sleep(200 * time.Millisecond)

	const n = 5
	base := time.Now().UnixMilli()
	want := make(map[int64]bool, n)
	for i := range n {
		want[base+int64(i)] = true
	}
	for i := range n {
		postPoint(t, point{
			GpsID:     gpsID,
			Latitude:  6.5,
			Longitude: 3.3 + float64(i)/10000,
			Bearing:   90,
			Timestamp: base + int64(i),
		})
	}

	got := awaitPoints(t, live, want, 15*time.Second)
	if len(got) != n {
		t.Fatalf("surviving client got %d points, want %d", len(got), n)
	}
	for i := range got {
		if got[i].Timestamp != base+int64(i) {
			t.Fatalf("point %d out of order: timestamp %d, want %d", i, got[i].Timestamp, base+int64(i))
		}
	}

	// The hub must still accept connections after the failed write.
	done := make(chan struct{})
	go func() {
		defer close(done)
		dial(t, "client-C")
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("hub wedged: could not connect a new client after a disconnect")
	}
}

// The simulator should be writing positions of its own accord.
func TestSimulatorIsProducingPoints(t *testing.T) {
	requireStack(t)

	count := func() int {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(apiURL() + "/api/v1/gps-points")
		if err != nil {
			t.Fatalf("list points: %v", err)
		}
		defer resp.Body.Close()
		var envelope struct {
			Data []point `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			t.Fatalf("decode points: %v", err)
		}
		return len(envelope.Data)
	}

	before := count()
	time.Sleep(12 * time.Second) // vehicles hop every 1-9s
	after := count()

	if after <= before {
		t.Fatalf("no new points in 12s (%d -> %d); is the simulator running?", before, after)
	}
	fmt.Printf("simulator produced %d points in 12s\n", after-before)
}
