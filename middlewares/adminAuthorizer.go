package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/debarkamondal/adda-cafe-backend/types"
)

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

func AdminAuthorizer(next HandleFunc) HandleFunc {
	var dbClient = dynamodb.NewFromConfig(cfg)
	return func(w http.ResponseWriter, r *http.Request) {
		csrfToken := r.Header.Get("X-CSRF-TOKEN")
		sessionToken, sesErr := r.Cookie("session_token")
		if sesErr != nil || csrfToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			body := map[string]any{"message": "Unauthorized"}
			json.NewEncoder(w).Encode(body)
			return
		}

		res, err := dbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
			TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
			Key: map[string]awsTypes.AttributeValue{
				"pk": &awsTypes.AttributeValueMemberS{Value: "session:backend"},
				"sk": &awsTypes.AttributeValueMemberS{Value: sessionToken.Value},
			},
		})
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			body := map[string]any{"message": "Error fetching session"}
			json.NewEncoder(w).Encode(body)
			return
		}
		if res.Item == nil {
			w.WriteHeader(http.StatusInternalServerError)
			body := map[string]any{"message": "Session not found"}
			json.NewEncoder(w).Encode(body)
			return
		}

		var session types.BackendSession
		err = attributevalue.UnmarshalMap(res.Item, &session)
		if session.CsrfToken != csrfToken || session.Role != types.AdminUser {
			w.WriteHeader(http.StatusBadRequest)
			body := map[string]any{"message": "Unauthorized"}
			json.NewEncoder(w).Encode(body)
			return
		}
		ctx := context.WithValue(r.Context(), types.SessionContextKey("session"), session)
		next(w, r.WithContext(ctx))
	}
}
