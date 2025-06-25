package menu

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	localType "github.com/debarkamondal/adda-cafe-backend/types"
)

func Patch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	dbClient := dynamodb.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(s3Client)
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

	var presignedUrl string
	var item localType.Product
	var updatedProduct localType.Product
	attributevalue.UnmarshalMap(res.Item, &item)
	json.NewDecoder(r.Body).Decode(&updatedProduct)

	//Update item fields
	if updatedProduct.Description != "" {
		item.Description = updatedProduct.Description
	}
	if updatedProduct.Price != 0 {
		item.Price = updatedProduct.Price
	}
	if updatedProduct.Title != "" {
		item.Title = updatedProduct.Title
	}
	if updatedProduct.Image != "" {
		item.Image = updatedProduct.Image
		imageSlice := strings.Split(updatedProduct.Image, ".")
		url, err := presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:      aws.String(os.Getenv("S3_BUCKET_NAME")),
			Key:         aws.String("items/" + updatedProduct.Image),
			ContentType: aws.String("image/" + imageSlice[1]),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(60 * int64(time.Second))
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			body := map[string]any{"message": "S3 error"}
			json.NewEncoder(w).Encode(body)
			return
		}
		presignedUrl = url.URL

	}
	body := map[string]any{
		"url": presignedUrl,
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)

}
