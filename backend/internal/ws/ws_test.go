package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
)

func newTestHub(t *testing.T) (*Hub, string) {
	t.Helper()
	hub := NewHub()
	handler := NewWsHandler(hub)
	srv := httptest.NewServer(http.HandlerFunc(handler.handleConnection))
	t.Cleanup(srv.Close)
	return hub, "ws" + strings.TrimPrefix(srv.URL, "http")
}

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

// waitForClients polls rather than sleeping, since registration happens on the
// server goroutine after Dial returns.
func waitForClients(t *testing.T, hub *Hub, want int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if hub.count() == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d clients, hub has %d", want, hub.count())
}

func point(id int) gpspoints.CreateGpsPoint {
	return gpspoints.CreateGpsPoint{GpsID: id, Latitude: 6.5, Longitude: 3.3, Timestamp: int64(id)}
}

// A failed write to one client must not wedge the hub. This is the regression
// test for the deadlock where Broadcast called Remove while holding the mutex.
func TestFailedWriteDoesNotWedgeHub(t *testing.T) {
	hub, url := newTestHub(t)

	dead := dial(t, url)
	waitForClients(t, hub, 1)
	dead.Close()

	// Push enough traffic that the writer goroutine certainly attempts a write.
	for i := range 10 {
		hub.Broadcast(point(i))
	}

	// The dead client should be reaped, and the hub must still be usable.
	deadline := time.Now().Add(3 * time.Second)
	for hub.count() != 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if hub.count() != 0 {
		t.Fatalf("dead client was not removed, hub has %d", hub.count())
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		dial(t, url)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("hub deadlocked: could not accept a new client after a failed write")
	}
	waitForClients(t, hub, 1)
}

// One dead client must not stop a healthy one from receiving positions.
func TestDeadClientDoesNotStarveHealthyOne(t *testing.T) {
	hub, url := newTestHub(t)

	dead := dial(t, url)
	live := dial(t, url)
	waitForClients(t, hub, 2)
	dead.Close()

	hub.Broadcast(point(42))

	if err := live.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	var got gpspoints.CreateGpsPoint
	if err := live.ReadJSON(&got); err != nil {
		t.Fatalf("healthy client received nothing: %v", err)
	}
	if got.GpsID != 42 {
		t.Fatalf("got GpsID %d, want 42", got.GpsID)
	}
}

// Positions must arrive in the order they were broadcast. The goroutine-per-write
// approach broke this; a single writer per connection guarantees it.
func TestOrderingIsPreserved(t *testing.T) {
	hub, url := newTestHub(t)
	c := dial(t, url)
	waitForClients(t, hub, 1)

	const n = 50
	for i := range n {
		hub.Broadcast(point(i))
	}

	if err := c.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	for i := range n {
		var got gpspoints.CreateGpsPoint
		if err := c.ReadJSON(&got); err != nil {
			t.Fatalf("read %d: %v", i, err)
		}
		if got.GpsID != i {
			t.Fatalf("message %d out of order: got GpsID %d", i, got.GpsID)
		}
	}
}

// Broadcast must not block on I/O, so it stays fast regardless of client state.
func TestBroadcastDoesNotBlock(t *testing.T) {
	hub, url := newTestHub(t)
	for range 5 {
		dial(t, url)
	}
	waitForClients(t, hub, 5)

	start := time.Now()
	for i := range 500 {
		hub.Broadcast(point(i))
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("500 broadcasts took %v; Broadcast is blocking on I/O", elapsed)
	}
}

// Connects, disconnects and broadcasts all at once. Run under -race, this is the
// regression test for guarding the client map with two different mutexes.
func TestConcurrentAddRemoveBroadcast(t *testing.T) {
	hub, url := newTestHub(t)

	stop := make(chan struct{})
	broadcasterDone := make(chan struct{})

	go func() {
		defer close(broadcasterDone)
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
				hub.Broadcast(point(i))
			}
		}
	}()

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 15 {
				c, _, err := websocket.DefaultDialer.Dial(url, nil)
				if err != nil {
					t.Error(err)
					return
				}
				c.Close()
			}
		}()
	}

	wg.Wait()
	close(stop)
	<-broadcasterDone
}
