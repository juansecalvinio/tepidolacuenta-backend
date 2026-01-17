package pkg

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Client represents a WebSocket client connection
type Client struct {
	ID           string
	UserID       string
	RestaurantID primitive.ObjectID
	Conn         *websocket.Conn
	Send         chan []byte
}

// Hub maintains active WebSocket connections and broadcasts messages
type Hub struct {
	// Registered clients grouped by restaurant ID
	clients map[primitive.ObjectID]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to clients of a specific restaurant
	broadcast chan *BroadcastMessage

	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast to a restaurant
type BroadcastMessage struct {
	RestaurantID primitive.ObjectID
	Data         interface{}
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[primitive.ObjectID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.RestaurantID] == nil {
				h.clients[client.RestaurantID] = make(map[*Client]bool)
			}
			h.clients[client.RestaurantID][client] = true
			h.mu.Unlock()
			log.Printf("Client %s registered for restaurant %s", client.ID, client.RestaurantID.Hex())

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.RestaurantID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					log.Printf("Client %s unregistered from restaurant %s", client.ID, client.RestaurantID.Hex())

					// Clean up empty restaurant groups
					if len(clients) == 0 {
						delete(h.clients, client.RestaurantID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.RestaurantID]
			h.mu.RUnlock()

			jsonData, err := json.Marshal(message.Data)
			if err != nil {
				log.Printf("Error marshaling broadcast message: %v", err)
				continue
			}

			for client := range clients {
				select {
				case client.Send <- jsonData:
				default:
					// Client's send channel is full, close connection
					h.mu.Lock()
					close(client.Send)
					delete(clients, client)
					h.mu.Unlock()
				}
			}
			log.Printf("Broadcasted message to %d clients of restaurant %s", len(clients), message.RestaurantID.Hex())
		}
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all clients of a restaurant
func (h *Hub) Broadcast(restaurantID primitive.ObjectID, data interface{}) {
	h.broadcast <- &BroadcastMessage{
		RestaurantID: restaurantID,
		Data:         data,
	}
}

// GetClientCount returns the number of connected clients for a restaurant
func (h *Hub) GetClientCount(restaurantID primitive.ObjectID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[restaurantID])
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister(c)
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		// We don't expect clients to send messages, only receive
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Send
		if !ok {
			// Hub closed the channel
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			return
		}
	}
}
