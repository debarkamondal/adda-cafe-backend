package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

func WsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	var dbClient = dynamodb.NewFromConfig(cfg)
	sessionToken, err := r.Cookie("session_token")
	csrfToken := r.URL.Query().Get("X-CSRF-TOKEN")
	if err != nil || csrfToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		body := map[string]any{"message": "Unauthorized"}
		json.NewEncoder(w).Encode(body)
		return
	}

	res, err := dbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]awsTypes.AttributeValue{
			"pk": &awsTypes.AttributeValueMemberS{Value: "session:backend"},
			"sk": &awsTypes.AttributeValueMemberS{Value: sessionToken.Value},
		},
	})
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Error fetching session"}
		json.NewEncoder(w).Encode(body)
		return
	}
	if res.Item == nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Session not found"}
		json.NewEncoder(w).Encode(body)
		return
	}

	var session types.BackendSession
	err = attributevalue.UnmarshalMap(res.Item, &session)
	if session.CsrfToken != csrfToken || session.Role != types.AdminUser {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Unauthorized"}
		json.NewEncoder(w).Encode(body)
		return
	}
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
