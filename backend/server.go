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
}

// Connected clients
var clients = make(map[*Client]bool)
var mutex   = &sync.Mutex{}

var broadcast = make(chan []byte)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

    clients[&Client{conn}] = true
    fmt.Println("New client connected, total clients:", len(clients))

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
            mutex.Lock()
            delete(clients, &Client{conn})
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
		for client := range clients {
            err := client.conn.WriteMessage(websocket.TextMessage, msg)
            if err != nil {
                log.Println("Write error:", err)
                client.conn.Close()
				delete(clients, client)
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
