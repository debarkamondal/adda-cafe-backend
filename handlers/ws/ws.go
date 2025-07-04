package ws

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/debarkamondal/adda-cafe-backend/types"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // Connected clients
var Broadcast = make(chan types.Order)       // Broadcast channel
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
	Mutex.Lock()
	clients[conn] = true
	Mutex.Unlock()
	for true {
	}
}
func HandleBroadcast() {
	for {
		// Grab the next order from the broadcast channel
		order := <-Broadcast

		// Send the message to all connected clients
		Mutex.Lock()
		for client := range clients {
			client.SetWriteDeadline(time.Now().Add(time.Second * 3))
			err := client.WriteJSON(order)
			if err != nil {
				fmt.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
		Mutex.Unlock()
	}
}
