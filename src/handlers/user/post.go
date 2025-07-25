package user

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/debarkamondal/adda-cafe-backend/src/types"
	"golang.org/x/crypto/bcrypt"
)

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
var dbClient = dynamodb.NewFromConfig(cfg)

func CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Id       string `json:"id"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid Token"}
		json.NewEncoder(w).Encode(body)
		return
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 12)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal Server Error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	user := &types.User{
		Pk:             "user",
		Sk:             creds.Id,
		HashedPassword: hashedPass,
		Role:           types.AdminUser,
	}
	marshalledUser, err := attributevalue.MarshalMap(user)
	_, err = dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Item:      marshalledUser,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal Server Error"}
		json.NewEncoder(w).Encode(body)
		return
	}

	w.WriteHeader(http.StatusOK)
	body := map[string]any{"message": "User created"}
	json.NewEncoder(w).Encode(body)

}
