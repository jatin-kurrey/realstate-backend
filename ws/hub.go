package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, refine this
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients by UserID.
	clients map[uint][]*Client
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
	// Mutex for concurrent access to clients
	mu sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = append(h.clients[client.UserID], client)
			h.mu.Unlock()
			log.Printf("User %d connected", client.UserID)
		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				for i, c := range clients {
					if c == client {
						h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
						close(client.Send)
						break
					}
				}
				if len(h.clients[client.UserID]) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			h.mu.Unlock()
			log.Printf("User %d disconnected", client.UserID)
		}
	}
}

// BroadcastToUser sends a message to all active connections of a specific user.
func (h *Hub) BroadcastToUser(userID uint, message interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	payload, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	if clients, ok := h.clients[userID]; ok {
		for _, client := range clients {
			select {
			case client.Send <- payload:
			default:
				// If send buffer is full, something is wrong with the client connection
				// We'll let the unregister handle it eventually
			}
		}
	}
}

// BroadcastToThread sends a message to all participants in a thread
func (h *Hub) BroadcastToThread(threadID uint, participants []uint, message interface{}) {
	for _, userID := range participants {
		h.BroadcastToUser(userID, message)
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// We don't really need to read anything from the client for now,
		// but we keep the connection open and responsive to pings.
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
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
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID uint) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to websocket: %v", err)
		return
	}

	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}
	client.Hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
