package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type StreamGenerator func(streamID string, broadcast chan<- Message) // Function type for creating stream handlers

type Hub struct {
	// Registered clients for each stream
	clients map[string]map[*Client]bool

	// Channel for broadcasting messages
	broadcast chan Message

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Active stream generators
	activeStreams map[string]chan struct{}

	// Function to create new stream generators
	streamFactory StreamGenerator

	mu sync.RWMutex
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	streamID string
}

type Message struct {
	streamID string
	data     []byte
}

func newHub(streamFactory StreamGenerator) *Hub {
	return &Hub{
		broadcast:     make(chan Message),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[string]map[*Client]bool),
		activeStreams: make(map[string]chan struct{}),
		streamFactory: streamFactory,
	}
}

func (h *Hub) activateStream(streamID string) {
	h.mu.Lock()
	if _, exists := h.activeStreams[streamID]; !exists {
		// Create stop channel for this stream
		stopCh := make(chan struct{})
		h.activeStreams[streamID] = stopCh

		// Start the stream generator in a new goroutine
		go h.streamFactory(streamID, h.broadcast)

		log.Printf("Activated stream: %s", streamID)
	}
	h.mu.Unlock()
}

func (h *Hub) deactivateStream(streamID string) {
	h.mu.Lock()
	if stopCh, exists := h.activeStreams[streamID]; exists {
		close(stopCh)
		delete(h.activeStreams, streamID)
		log.Printf("Deactivated stream: %s", streamID)
	}
	h.mu.Unlock()
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.streamID]; !ok {
				h.clients[client.streamID] = make(map[*Client]bool)
				// Activate stream for first client
				go h.activateStream(client.streamID)
			}
			h.clients[client.streamID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.streamID]; ok {
				delete(h.clients[client.streamID], client)
				close(client.send)

				// If no more clients for this stream, deactivate it
				if len(h.clients[client.streamID]) == 0 {
					go h.deactivateStream(client.streamID)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[message.streamID]; ok {
				for client := range clients {
					select {
					case client.send <- message.data:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Example usage in main:
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, implement proper origin checking
	},
}

// serveWs handles websocket requests from clients
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Get the requested stream ID from URL parameters
	streamID := r.URL.Query().Get("stream")
	if streamID == "" {
		log.Println("No stream ID provided")
		conn.Close()
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		streamID: streamID,
	}

	// Register this client with the hub
	client.hub.register <- client

	// Start the write pump in a new goroutine
	go func() {
		defer func() {
			client.hub.unregister <- client
			conn.Close()
		}()

		for {
			select {
			case message, ok := <-client.send:
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(message)

				if err := w.Close(); err != nil {
					return
				}
			}
		}
	}()
}

func main() {
	// Define your stream generator factory
	streamFactory := func(streamID string, broadcast chan<- Message) {
		// This is where you would implement your custom data generation logic
		// Example:
		// - Parse streamID for parameters
		// - Connect to data sources
		// - Generate data
		// - Send data through broadcast channel
		// - Monitor stopCh for shutdown signal
	}

	hub := newHub(streamFactory)
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
