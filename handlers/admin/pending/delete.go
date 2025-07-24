package pending

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
var dbClient = dynamodb.NewFromConfig(cfg)

type Body struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

func Delete(w http.ResponseWriter, r *http.Request) {
	var req Body
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

	_, err = dbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
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
