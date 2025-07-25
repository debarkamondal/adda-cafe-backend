package pending

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/debarkamondal/adda-cafe-backend/src/clients"
	localTypes "github.com/debarkamondal/adda-cafe-backend/src/types"
)

type contextKey string

func Put(w http.ResponseWriter, r *http.Request) {
	var req Body
	ctx := r.Context()
	session, _ := ctx.Value(localTypes.SessionContextKey("session")).(localTypes.BackendSession)
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

	pending := &localTypes.FlagAction{
		Pk:        "flagged",
		Sk:        req.Id,
		Blame:     session.Name,
		Reason:    req.Reason,
		Type:      actionType,
		CreatedAt: time.Now().UnixMilli(),
	}

	marshalledData, err := attributevalue.MarshalMap(pending)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal Server Error."}
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
				Put: &types.Put{
					TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
					Item:      marshalledData,
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
	w.WriteHeader(http.StatusNoContent)

}
