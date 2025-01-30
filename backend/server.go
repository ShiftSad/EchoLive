package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
    "sync"
)

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        // Adjust to your needs; in production, be sure to verify the origin.
        return true
    },
}

// Client represents a single WebSocket connection
type Client struct {
    conn *websocket.Conn
    role string
}

// We maintain separate sets for broadcasters and listeners.
// Using sync.Mutex to protect these sets during concurrent reads/writes.
var (
    broadcasters = make(map[*Client]bool)
    listeners    = make(map[*Client]bool)
    mu           sync.Mutex
)

func main() {
    http.HandleFunc("/ws", handleConnections)

    fmt.Println("Server running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleConnections upgrades the HTTP connection to a WebSocket
// and adds the client to the appropriate set (broadcasters or listeners).
func handleConnections(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket Upgrade Error:", err)
        return
    }

    // Ensure the connection is closed when this function returns
    defer conn.Close()

    // Determine role from query param: ?mode=broadcast or ?mode=listen
    role := r.URL.Query().Get("mode")
    client := &Client{conn: conn, role: role}

    mu.Lock()
    switch role {
    case "broadcast":
        broadcasters[client] = true
        fmt.Println("Broadcaster connected")
    case "listen":
        listeners[client] = true
        fmt.Println("Listener connected")
    default:
        // If neither role is provided, you might decide to close or handle differently.
        // We'll treat them as a listener by default or just disconnect:
        listeners[client] = true
        fmt.Println("Unspecified role, treating as listener")
    }
    mu.Unlock()

    // Read loop for messages from this client
    for {
        mt, message, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Client (%s) disconnected: %v", role, err)
            removeClient(client)
            break
        }

        // Route messages based on the client's role
        if role == "broadcast" {
            // Forward the broadcaster's message to all listeners
            mu.Lock()
            for listener := range listeners {
                // Forward the exact message
                if err := listener.conn.WriteMessage(mt, message); err != nil {
                    log.Println("Error writing to listener:", err)
                }
            }
            mu.Unlock()
        } else {
            // Forward the listener's message to all broadcasters
            mu.Lock()
            for broadcaster := range broadcasters {
                if err := broadcaster.conn.WriteMessage(mt, message); err != nil {
                    log.Println("Error writing to broadcaster:", err)
                }
            }
            mu.Unlock()
        }
    }
}

// removeClient removes the given client from whichever map it belongs to
func removeClient(c *Client) {
    mu.Lock()
    defer mu.Unlock()

    switch c.role {
    case "broadcast":
        delete(broadcasters, c)
    case "listen":
        delete(listeners, c)
    default:
        delete(listeners, c)
    }
}
