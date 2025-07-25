package menu

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/debarkamondal/adda-cafe-backend/src/types"
)

var dbClient = dynamodb.NewFromConfig(cfg)
var s3Client = s3.NewFromConfig(cfg)

func Post(w http.ResponseWriter, r *http.Request) {
	var product types.Product
	presigner := s3.NewPresignClient(s3Client)
	id, err := uuid.NewV7()
	json.NewDecoder(r.Body).Decode(&product)
	product.Image = id.String() + "." + strings.Split(product.Image, ".")[1]

	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	product.Pk = "menu"
	product.Sk = id.String()

	item, err := attributevalue.MarshalMap(product)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid input"}
		json.NewEncoder(w).Encode(body)
		return
	}
	_, err = dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Item:      item,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	url, err := presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("S3_BUCKET_NAME")),
		Key:         aws.String("menu/" + product.Image),
		ContentType: aws.String("image/" + strings.Split(product.Image, ".")[1]),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(120 * int64(time.Second))
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "S3 error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	res := map[string]any{
		"url": url.URL,
		"id":  id,
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
