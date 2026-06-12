package gps

import "sync"

type EventHub struct {
	mu        sync.Mutex
	listeners map[chan GpsResponse]struct{}
}

func NewEventHub() *EventHub {
	return &EventHub{
		listeners: make(map[chan GpsResponse]struct{}),
	}
}

func (h *EventHub) Subscribe() chan GpsResponse {
	ch := make(chan GpsResponse, 4)
	h.mu.Lock()
	h.listeners[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *EventHub) Unsubscribe(ch chan GpsResponse) {
	h.mu.Lock()
	delete(h.listeners, ch)
	h.mu.Unlock()
}

func (h *EventHub) Publish(g GpsResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.listeners {
		select {
		case ch <- g:
		default:
		}
	}
}
