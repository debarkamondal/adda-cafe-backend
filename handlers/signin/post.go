package signin

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	localTypes "github.com/debarkamondal/adda-cafe-backend/types"
)


type credentials struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
var dbClient = dynamodb.NewFromConfig(cfg)

func Post(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		body := map[string]any{"message": "Missing credentials"}
		json.NewEncoder(w).Encode(body)
		return
	}
	res, err := dbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "user"},
			"sk": &types.AttributeValueMemberS{Value: creds.Id},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal Server Error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	if res.Item == nil {
		w.WriteHeader(http.StatusUnauthorized)
		body := map[string]any{"message": "User invalid"}
		json.NewEncoder(w).Encode(body)
		return
	}

	var user localTypes.User
	attributevalue.UnmarshalMap(res.Item, &user)

	uid, err := uuid.NewV7()
	csrf, err := uuid.NewV7()
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(creds.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		body := map[string]any{"message": "Unauthorized"}
		json.NewEncoder(w).Encode(body)
		return
	}

	currentTime := time.Now().UnixMilli()

	session := localTypes.BackendSession{
		Pk:        "session",
		Sk:        uid.String(),
		Role:      user.Role,
		CsrfToken: csrf.String(),
		CreatedAt: currentTime,
	}

	marshalledSession, err := attributevalue.MarshalMap(session)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal Server Error"}
		json.NewEncoder(w).Encode(body)
		return
	}

	_, err = dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Item:      marshalledSession,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal Server Error"}
		json.NewEncoder(w).Encode(body)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    uid.String(),
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrf.String(),
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
	body := map[string]any{"message": "Login success"}
	json.NewEncoder(w).Encode(body)
}
