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
	}
	defer conn.Close()

	query := r.URL.Query()
	mode := query.Get("mode") // broadcaster or listener

	hub.mu.Lock()
	if mode == "broadcast" {
		hub.broadcasters[conn] = true
	} else if mode == "listen" {
		hub.listeners[conn] = true
	}
	hub.mu.Unlock()

	// read and forward audio only if broadcaster
	if mode == "broadcast" {
		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				break
			}
			fmt.Println("Received audio from broadcaster")
			hub.broadcast <- p // forward audio to listeners
		}

		// Remove broadcaster from hub
		hub.mu.Lock()
		delete(hub.broadcasters, conn)
		hub.mu.Unlock()
	}

	// Remove listener on disconnect
	hub.mu.Lock()
	delete(hub.listeners, conn)
	hub.mu.Unlock()
}

// Broadcast received audio to opted-in listeners
func handleBroadcasts() {
	for {
		msg := <-hub.broadcast
		hub.mu.Lock()
		for conn := range hub.listeners { // Send only to listeners
			err := conn.WriteMessage(websocket.BinaryMessage, msg)
			if err != nil {
				conn.Close()
				delete(hub.listeners, conn)
			}
		}
		hub.mu.Unlock()
	}
}
