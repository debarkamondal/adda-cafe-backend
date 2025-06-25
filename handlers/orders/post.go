package orders

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/debarkamondal/adda-cafe-backend/handlers/ws"
	"github.com/debarkamondal/adda-cafe-backend/types"
	"github.com/google/uuid"
)

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

func Post(w http.ResponseWriter, r *http.Request) {
	var dbClient = dynamodb.NewFromConfig(cfg)
	var order types.Order

	sessionToken, _ := r.Cookie("session_token") //This error is already handeled in the UserAuthMiddleware

	json.NewDecoder(r.Body).Decode(&order)
	if len(order.Items) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Empty item list"}
		json.NewEncoder(w).Encode(body)
		return
	}

	id, err := uuid.NewV7()
	currentTime := time.Now().UnixMilli()

	order.Pk = "order"
	order.Sk = id.String()
	order.SessionId = sessionToken.Value
	order.CreatedAt = currentTime

	marshalledOrder, err := attributevalue.MarshalMap(order)
	ws.Broadcast <- order

	_, err = dbClient.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: []awsTypes.TransactWriteItem{
			{
				Update: &awsTypes.Update{
					TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
					Key: map[string]awsTypes.AttributeValue{
						"pk": &awsTypes.AttributeValueMemberS{Value: "session"},          // Partition Key
						"sk": &awsTypes.AttributeValueMemberS{Value: sessionToken.Value}, // Sort Key
					},
					UpdateExpression: aws.String("SET #orders = list_append(#orders, :order), updatedAt = :updatedAt"),
					ExpressionAttributeNames: map[string]string{
						"#orders": "orders",
					},
					ExpressionAttributeValues: map[string]awsTypes.AttributeValue{
						":order": &awsTypes.AttributeValueMemberL{
							Value: []awsTypes.AttributeValue{
								&awsTypes.AttributeValueMemberS{Value: id.String()},
							},
						},
						":updatedAt": &awsTypes.AttributeValueMemberN{Value: strconv.FormatInt(currentTime, 10)},
					},
				},
			},
			{
				Put: &awsTypes.Put{
					TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
					Item:      marshalledOrder,
				},
			},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	w.WriteHeader(http.StatusOK)
	body := map[string]any{"message": "Order created"}
	json.NewEncoder(w).Encode(body)
}
