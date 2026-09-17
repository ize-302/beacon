package vehicles

import "sync"

// EventHub is used to notify subscribers (the simulator, over SSE) when a new vehicle is
// created.
type EventHub struct {
	mu        sync.Mutex
	listeners map[chan VehicleResponse]struct{}
}

func NewEventHub() *EventHub {
	return &EventHub{
		listeners: make(map[chan VehicleResponse]struct{}),
	}
}

func (h *EventHub) Subscribe() chan VehicleResponse {
	ch := make(chan VehicleResponse, 4)
	h.mu.Lock()
	h.listeners[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *EventHub) Unsubscribe(ch chan VehicleResponse) {
	h.mu.Lock()
	delete(h.listeners, ch)
	h.mu.Unlock()
}

func (h *EventHub) Publish(g VehicleResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.listeners {
		select {
		case ch <- g:
		default:
		}
	}
}
