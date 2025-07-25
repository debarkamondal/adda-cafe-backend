package pending

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/debarkamondal/adda-cafe-backend/src/clients"
	localType "github.com/debarkamondal/adda-cafe-backend/src/types"
)

type Body struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

func Delete(w http.ResponseWriter, r *http.Request) {
	var req Body
	session, _ := r.Context().Value(localType.SessionContextKey("session")).(localType.BackendSession)
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid request"}
		json.NewEncoder(w).Encode(body)
		return
	}

	var actionType string
	switch req.Type {
	case "session":
		actionType = "session"
	case "order":
		actionType = "order"
	default:
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid request"}
		json.NewEncoder(w).Encode(body)
		return

	}

	_, err = clients.DBClient.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Delete: &types.Delete{
					TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
					Key: map[string]types.AttributeValue{
						"pk": &types.AttributeValueMemberS{Value: "pending"},
						"sk": &types.AttributeValueMemberS{Value: actionType + ":" + req.Id},
					},
				},
			},
			{
				Update: &types.Update{
					TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
					Key: map[string]types.AttributeValue{
						"pk": &types.AttributeValueMemberS{Value: "session"},
						"sk": &types.AttributeValueMemberS{Value: req.Id},
					},
					UpdateExpression: aws.String("SET blame = :acceptedBy"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":acceptedBy": &types.AttributeValueMemberS{Value: session.Name},
					},
				},
			},
		},
	})
	_, err = clients.DBClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "pending"},
			"sk": &types.AttributeValueMemberS{Value: actionType + ":" + req.Id},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
