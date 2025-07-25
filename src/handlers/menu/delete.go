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
	"github.com/aws/aws-sdk-go-v2/service/s3"

	localType "github.com/debarkamondal/adda-cafe-backend/src/types"
)

var test localType.Product

func Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	dbClient := dynamodb.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	res, err := dbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "item"}, // Partition Key
			"sk": &types.AttributeValueMemberS{Value: id},     // Sort Key
		},
	})
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	if res.Item == nil {

		w.WriteHeader(http.StatusNotFound)
		body := map[string]any{"message": "Requested item not found"}
		json.NewEncoder(w).Encode(body)
		return
	}
	var item localType.Product
	attributevalue.UnmarshalMap(res.Item, &item)
	_, deleteErr := dbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "item"}, // Partition Key
			"sk": &types.AttributeValueMemberS{Value: id},     // Sort Key
		},
	})
	if deleteErr != nil {

		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	_, s3Err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
		Key:    aws.String("items/" + item.Image),
	})
	if s3Err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "S3 error"}
		json.NewEncoder(w).Encode(body)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	body := map[string]any{"message": "Item deleted."}
	json.NewEncoder(w).Encode(body)
}
