package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// Event is the JSON envelope sent to WebSocket clients.
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// Hub maintains the set of active clients and broadcasts events to them.
type Hub struct {
	// NotifyChan receives a signal every time a message is broadcast.
	// Agents can select on this channel to wake immediately.
	NotifyChan chan struct{}

	mu         sync.RWMutex
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	done       chan struct{}
}

// NewHub creates a new Hub. Call Run() to start processing.
func NewHub() *Hub {
	return &Hub{
		NotifyChan: make(chan struct{}, 1),
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event, 256),
		done:       make(chan struct{}),
	}
}

// Run starts the hub's event loop. It blocks until Stop is called.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			data, err := json.Marshal(event)
			if err != nil {
				slog.Error("ws hub: marshal event", "error", err)
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					// Client buffer full — drop and disconnect.
					go h.removeClient(client)
				}
			}
			h.mu.RUnlock()

			// Signal NotifyChan (non-blocking).
			select {
			case h.NotifyChan <- struct{}{}:
			default:
			}

		case <-h.done:
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return
		}
	}
}

// Stop shuts down the hub's event loop.
func (h *Hub) Stop() {
	close(h.done)
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// Broadcast sends an event to all connected clients.
func (h *Hub) Broadcast(event Event) {
	select {
	case h.broadcast <- event:
	default:
		slog.Warn("ws hub: broadcast channel full, dropping event")
	}
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) removeClient(c *Client) {
	h.unregister <- c
}
