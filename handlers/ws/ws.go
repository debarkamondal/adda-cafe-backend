package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // Connected clients
var Broadcast = make(chan map[string]any)            // Broadcast channel
var Mutex = &sync.Mutex{}                    // Protect clients map

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()
	// Listen for incoming messages
	Mutex.Lock()
	clients[conn] = true
	Mutex.Unlock()

	for {
		// _, message, err := conn.ReadMessage()
		// if err != nil {
		// 	Mutex.Lock()
		// 	delete(clients, conn)
		// 	Mutex.Unlock()
		// 	break
		// }
		// Broadcast <- message
	}
}
func HandleBroadcast() {
	for {
		// Grab the next message from the broadcast channel
		message := <-Broadcast

		// Send the message to all connected clients
		Mutex.Lock()
		for client := range clients {
			temp, err := json.Marshal(message)
			if err!= nil{
				fmt.Println("Error bad JSON")
				return
			}
			err = client.WriteMessage(websocket.TextMessage, temp)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		Mutex.Unlock()
	}
}
