package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/admin/ws"
	"github.com/debarkamondal/adda-cafe-backend/src/middlewares"
)

var cfg, err = config.LoadDefaultConfig(context.Background(), config.WithRegion("ap-south-1"))
var DBClient = dynamodb.NewFromConfig(cfg)

func main() {
	mux := InitRoutes(http.NewServeMux())

	go ws.HandleBroadcast()

	if os.Getenv("PROXY") == "true" {
		fmt.Println("Listening on port 8080 (proxied)")
		if err := http.ListenAndServe(":8080", http.StripPrefix(os.Getenv("URI_PREFIX"), mux)); err != nil {
			fmt.Println(err)
			fmt.Println("Couldn't initiate server on port 8080")
		}
	} else {
		fmt.Println("Listening on port 8080 (unproxied)")
		if err := http.ListenAndServe(":8080", http.StripPrefix(os.Getenv("URI_PREFIX"), middlewares.CORS(mux))); err != nil {
			fmt.Println(err)
			fmt.Println("Couldn't initiate server on port 8080")
		}
	}

}
