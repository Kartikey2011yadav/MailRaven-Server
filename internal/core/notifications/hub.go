package notifications

import (
	"sync"
)

// EventType definition
type EventType int

const (
	EventNewMessage EventType = iota
)

// Event represents a system event
type Event struct {
	Type    EventType
	UserID  string
	Mailbox string
	Data    interface{}
}

// Hub allows subscription to events
type Hub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event // Keyed by UserID
}

// Global Hub instance (for MVP)
var GlobalHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string][]chan Event),
	}
}

// Subscribe adds a channel to receive events for a user
func (h *Hub) Subscribe(userID string) chan Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Event, 10)
	h.subscribers[userID] = append(h.subscribers[userID], ch)
	return ch
}

// Unsubscribe removes a channel
func (h *Hub) Unsubscribe(userID string, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subs := h.subscribers[userID]
	for i, sub := range subs {
		if sub == ch {
			// Remove from slice
			h.subscribers[userID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	if len(h.subscribers[userID]) == 0 {
		delete(h.subscribers, userID)
	}
}

// Broadcast sends an event to all subscribers for a user
func (h *Hub) Broadcast(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs := h.subscribers[event.UserID]
	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// Non-blocking send, drop if full to avoid stalling
		}
	}
}
