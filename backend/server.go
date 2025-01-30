package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub to manage users
type Hub struct {
	mu           sync.Mutex
	broadcasters map[*websocket.Conn]bool
	listeners    map[*websocket.Conn]bool
	broadcast    chan []byte
}

var hub = Hub{
	broadcasters: make(map[*websocket.Conn]bool),
	listeners:    make(map[*websocket.Conn]bool),
	broadcast:    make(chan []byte),
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	go handleBroadcasts()

	port := ":8080"
	fmt.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
        return
    }
    defer conn.Close()

    query := r.URL.Query()
    mode := query.Get("mode")
    
    if mode == "broadcast" {
        handleBroadcaster(conn)
    } else if mode == "listen" {
        handleListener(conn)
    } else {
        log.Printf("Invalid mode: %s", mode)
        return
    }
}

func handleBroadcaster(conn *websocket.Conn) {
    hub.mu.Lock()
    hub.broadcasters[conn] = true
    hub.mu.Unlock()
    log.Printf("New broadcaster connected. Total broadcasters: %d", len(hub.broadcasters))

    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Broadcaster error: %v", err)
            break
        }
        log.Printf("Received message type: %d, size: %d bytes", messageType, len(p))
        hub.broadcast <- p
    }

    hub.mu.Lock()
    delete(hub.broadcasters, conn)
    hub.mu.Unlock()
}

func handleListener(conn *websocket.Conn) {
    hub.mu.Lock()
    hub.listeners[conn] = true
    listenerCount := len(hub.listeners)
    hub.mu.Unlock()
    
    log.Printf("New listener connected. Total listeners: %d", listenerCount)

    // Keep connection alive until client disconnects
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Listener disconnected: %v", err)
            hub.mu.Lock()
            delete(hub.listeners, conn)
            hub.mu.Unlock()
            return
        }
    }
}

// Broadcast received audio to opted-in listeners
func handleBroadcasts() {
    for {
        msg := <-hub.broadcast
        hub.mu.Lock()
        listenerCount := len(hub.listeners)
        log.Printf("Broadcasting %d bytes to %d listeners", len(msg), listenerCount)
        
        for conn := range hub.listeners {
            err := conn.WriteMessage(websocket.BinaryMessage, msg)
            if err != nil {
                log.Printf("Error sending to listener: %v", err)
                conn.Close()
                delete(hub.listeners, conn)
            }
        }
        hub.mu.Unlock()
    }
}