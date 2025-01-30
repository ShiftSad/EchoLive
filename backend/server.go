package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
    conn *websocket.Conn
    role string
}

// Connected clients
var broadcasters = make(map[*Client]bool)
var listeners    = make(map[*Client]bool)
var mutex = &sync.Mutex{}

var broadcast = make(chan []byte)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	role := r.URL.Query().Get("role")
    switch role {
    case "broadcast":
        broadcasters[&Client{conn, role}] = true
        fmt.Println("New broadcaster connected")
    case "listen":
        listeners[&Client{conn, role}] = true
        fmt.Println("New listener connected")
    case "both":
        broadcasters[&Client{conn, "broadcast"}] = true
        listeners[&Client{conn, "listen"}] = true
        fmt.Println("New broadcaster and listener connected")
    }

	for {
        // Only read messages from broadcasters
        if role != "broadcast" {
            continue
        }

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
            mutex.Lock()
            delete(broadcasters, &Client{conn, role})
            delete(listeners, &Client{conn, role})
            mutex.Unlock()
            break
		}

		// Broadcast message to all clients
        broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
        fmt.Println("Broadcasting message:", string(msg))
		// Send message to all connected clients
		mutex.Lock()
		for listener := range listeners {
            err := listener.conn.WriteMessage(websocket.TextMessage, msg)
            if err != nil {
                log.Println("Write error:", err)
                listener.conn.Close()
				delete(listeners, listener)
            }
		}
		mutex.Unlock()
	}
}

func main() {
	// Serve WebSocket on /ws route
	http.HandleFunc("/ws", handleConnections)

	// Start message handling in a goroutine
	go handleMessages()

	port := ":8080"
	fmt.Println("WebSocket server running on ws://localhost" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
