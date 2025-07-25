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
	"github.com/debarkamondal/adda-cafe-backend/src/clients"
	"github.com/debarkamondal/adda-cafe-backend/src/types"
	"github.com/gorilla/websocket"
)

var wsClients = make(map[*websocket.Conn]bool) // Connected clients
var Broadcast = make(chan any)                 // Broadcast channel
var Mutex = &sync.Mutex{}                      // Protect clients map

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

func WsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	sessionToken, err := r.Cookie("session_token")
	csrfToken := r.URL.Query().Get("X-CSRF-TOKEN")
	if err != nil || csrfToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		body := map[string]any{"message": "Unauthorized"}
		json.NewEncoder(w).Encode(body)
		return
	}

	res, err := clients.DBClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]awsTypes.AttributeValue{
			"pk": &awsTypes.AttributeValueMemberS{Value: "session:backend"},
			"sk": &awsTypes.AttributeValueMemberS{Value: sessionToken.Value},
		},
	})
	if err != nil {
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
	paRes, err := clients.DBClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("DB_TABLE_NAME")),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]awsTypes.AttributeValue{
			":pk": &awsTypes.AttributeValueMemberS{Value: "pending"},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Error fetching session"}
		json.NewEncoder(w).Encode(body)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	if len(paRes.Items) > 0 {
		var pendingActions []types.PendingAction
		err = attributevalue.UnmarshalListOfMaps(paRes.Items, &pendingActions)
		Broadcast <- pendingActions
	}

	defer conn.Close()
	Mutex.Lock()
	wsClients[conn] = true
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
		for client := range wsClients {
			client.SetWriteDeadline(time.Now().Add(time.Second * 3))
			err := client.WriteJSON(order)
			if err != nil {
				client.Close()
				delete(wsClients, client)
			}
		}
		Mutex.Unlock()
	}
}
