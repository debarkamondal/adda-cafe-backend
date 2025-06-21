package products

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	localType "github.com/debarkamondal/adda-cafe-backend/types"
)

func Get(w http.ResponseWriter, r *http.Request) {
	dbClient := dynamodb.NewFromConfig(cfg)
	res, err := dbClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("go-test"),
		KeyConditionExpression: aws.String("pk = :items"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":items": &types.AttributeValueMemberS{Value: "item"},
		},
	})
	if err != nil {
		fmt.Println(err)
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
