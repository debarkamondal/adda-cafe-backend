package products

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/debarkamondal/adda-cafe-backend/types"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var product types.Product
	dbClient := dynamodb.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(s3Client)
	id, err := uuid.NewV7()
	json.NewDecoder(r.Body).Decode(&product)

	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	product.Pk = "item"
	product.Sk = id.String()

	item, err := attributevalue.MarshalMap(product)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid input"}
		json.NewEncoder(w).Encode(body)
		return
	}
	_, err = dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("go-test"),
		Item:      item,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}

	imageSlice := strings.Split(product.Image, ".")
	url, err := presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String("dezire-golang-bucket"),
		Key:         aws.String("items/" + product.Image),
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
	res := map[string]any{
		"url": url.URL,
		"id":  id,
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
