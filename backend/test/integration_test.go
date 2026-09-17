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
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type point struct {
	VehicleID int     `json:"vehicle_id"`
	Bearing   float64 `json:"bearing"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

// frame is the WebSocket payload: one message per write, carrying every position
// recorded together.
type frame struct {
	Type   string  `json:"type"`
	Points []point `json:"points"`
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

// testVehicle creates a vehicle for the calling test and removes it afterwards.
// Deletes cascade to its points, so nothing is left behind.
func testVehicle(t *testing.T) int {
	t.Helper()
	return makeTrackedVehicle(t)
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
		var f frame
		if err := c.ReadJSON(&f); err != nil {
			t.Fatalf("expected %d more of our points, read failed: %v", remaining, err)
		}
		if f.Type != "positions" {
			t.Fatalf("unexpected frame type %q", f.Type)
		}
		for _, p := range f.Points {
			if want[p.Timestamp] {
				got = append(got, p)
				remaining--
			}
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
	vehicleID := testVehicle(t)

	c := dial(t, "client")
	time.Sleep(200 * time.Millisecond) // let the hub register the connection

	// A distinctive timestamp separates our points from the simulator's.
	base := time.Now().UnixMilli()
	want := map[int64]bool{base: true}
	postPoint(t, point{VehicleID: vehicleID, Latitude: 6.5, Longitude: 3.3, Bearing: 90, Timestamp: base})

	got := awaitPoints(t, c, want, 10*time.Second)
	if len(got) != 1 {
		t.Fatalf("got %d points, want 1", len(got))
	}
	if got[0].VehicleID != vehicleID {
		t.Fatalf("got VehicleID %d, want %d", got[0].VehicleID, vehicleID)
	}
}

// The regression test for the hub deadlock: one client vanishing must not stop
// delivery to the others, reorder anything, or stop new clients connecting.
func TestHubSurvivesClientDisconnect(t *testing.T) {
	requireStack(t)
	vehicleID := testVehicle(t)

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
			VehicleID: vehicleID,
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

// --- delete behaviour -------------------------------------------------------
//
// These live here rather than in a unit test because the thing under test is the
// foreign key between vehicles and gpspoints, which only a real database
// enforces.

func doJSON(t *testing.T, method, url string, body any) (int, []byte) {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, out
}

func idFrom(t *testing.T, body []byte) int {
	t.Helper()
	var envelope struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode id: %v (%s)", err, body)
	}
	return envelope.Data.ID
}

// makeTrackedVehicle creates a vehicle and removes it when the test finishes.
// One call is all it takes now — there is no device to register afterwards.
func makeTrackedVehicle(t *testing.T) int {
	t.Helper()
	plate := fmt.Sprintf("TEST-%d", time.Now().UnixNano()%1000000)

	code, body := doJSON(t, http.MethodPost, apiURL()+"/api/v1/vehicles", map[string]any{
		"plate_number": plate,
		"vehicle_type": "car",
		"device_sn":    plate + "-SN",
	})
	if code != http.StatusCreated {
		t.Fatalf("create vehicle returned %d: %s", code, body)
	}
	vehicleID := idFrom(t, body)

	// Delete cascades to the vehicle's points. A 404 just means the test already
	// deleted it, which is the whole point of some of these.
	t.Cleanup(func() {
		code, _ := doJSON(t, http.MethodDelete, fmt.Sprintf("%s/api/v1/vehicles/%d", apiURL(), vehicleID), nil)
		if code != http.StatusNoContent && code != http.StatusNotFound {
			t.Errorf("cleanup of vehicle %d returned %d", vehicleID, code)
		}
	})

	return vehicleID
}

func statusOf(t *testing.T, url string) int {
	t.Helper()
	code, _ := doJSON(t, http.MethodGet, url, nil)
	return code
}

// Deleting a vehicle must actually delete it, and take its history with it.
func TestDeleteVehicleCascades(t *testing.T) {
	requireStack(t)
	vehicleID := makeTrackedVehicle(t)

	postPoint(t, point{VehicleID: vehicleID, Latitude: 6.5, Longitude: 3.3, Bearing: 90, Timestamp: time.Now().UnixMilli()})

	// History exists before the delete.
	if code := statusOf(t, fmt.Sprintf("%s/api/v1/vehicles/%d/history", apiURL(), vehicleID)); code != http.StatusOK {
		t.Fatalf("history unavailable before delete (status %d)", code)
	}

	if code, body := doJSON(t, http.MethodDelete, fmt.Sprintf("%s/api/v1/vehicles/%d", apiURL(), vehicleID), nil); code != http.StatusNoContent {
		t.Fatalf("delete vehicle returned %d: %s", code, body)
	}

	// The old bug: the endpoint reported success while the row survived.
	if code := statusOf(t, fmt.Sprintf("%s/api/v1/vehicles/%d", apiURL(), vehicleID)); code != http.StatusNotFound {
		t.Fatalf("vehicle still readable after delete (status %d)", code)
	}
	// ON DELETE CASCADE should have taken the points with it.
	if code := statusOf(t, fmt.Sprintf("%s/api/v1/vehicles/%d/history", apiURL(), vehicleID)); code != http.StatusNotFound {
		t.Fatalf("history survived its vehicle being deleted (status %d)", code)
	}
}

// Deleting something that is not there must 404 rather than silently succeed.
func TestDeleteMissingReturns404(t *testing.T) {
	requireStack(t)

	if code, _ := doJSON(t, http.MethodDelete, apiURL()+"/api/v1/vehicles/99999999", nil); code != http.StatusNotFound {
		t.Fatalf("deleting a missing vehicle returned %d, want 404", code)
	}
}

// The headline of the merge: one call creates a vehicle and it starts moving,
// with no separate device to register.
func TestNewVehicleIsTrackedImmediately(t *testing.T) {
	requireStack(t)

	c := dial(t, "tracker")
	time.Sleep(200 * time.Millisecond)

	vehicleID := makeTrackedVehicle(t)

	// The simulator learns about it over SSE and starts posting positions, which
	// reach us over the socket without any further setup.
	deadline := time.Now().Add(30 * time.Second)
	if err := c.SetReadDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	for {
		var f frame
		if err := c.ReadJSON(&f); err != nil {
			t.Fatalf("no position for the new vehicle within the deadline: %v", err)
		}
		for _, p := range f.Points {
			if p.VehicleID == vehicleID {
				return // tracked, with one API call
			}
		}
	}
}
