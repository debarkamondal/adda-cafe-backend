package clients

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var DBClient *dynamodb.Client
var S3Client *s3.Client

func Init() {
	var cfg, err = config.LoadDefaultConfig(context.Background(), config.WithRegion("ap-south-1"))
	if err != nil {
		fmt.Println("Error initializing db client.")
	}
	DBClient = dynamodb.NewFromConfig(cfg)
	S3Client= s3.NewFromConfig(cfg)
}
