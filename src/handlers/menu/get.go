package menu

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/debarkamondal/adda-cafe-backend/src/clients"
	localType "github.com/debarkamondal/adda-cafe-backend/src/types"
)

func Get(w http.ResponseWriter, r *http.Request) {
	res, err := clients.DBClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("DB_TABLE_NAME")),
		KeyConditionExpression: aws.String("pk = :items"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":items": &types.AttributeValueMemberS{Value: "menu"},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}

	var items []localType.Product
	attributevalue.UnmarshalListOfMaps(res.Items, &items)
	if len(items) <= 0 {
		w.WriteHeader(http.StatusNotFound)
		body := map[string]any{"message": "Requested item not found"}
		json.NewEncoder(w).Encode(body)
		return
	}
	json.NewEncoder(w).Encode(items)

}
func GetById(w http.ResponseWriter, r *http.Request) {
	res, err := clients.DBClient.GetItem(r.Context(), &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "menu"},
			"sk": &types.AttributeValueMemberS{Value: r.PathValue("id")},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	var item localType.Product
	err = attributevalue.UnmarshalMap(res.Item, &item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}
