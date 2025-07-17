package login

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	localTypes "github.com/debarkamondal/adda-cafe-backend/types"
	"github.com/debarkamondal/adda-cafe-backend/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

type adminCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type adminData struct {
	Username       string `json:"username" dynamodbav:"pk"` //admin:<username>
	Sk             string `json:"sk" dynamodbav:"sk"`       //data
	HashedPassword string `json:"-" dynamodbav:"hashedPassword"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func Post(w http.ResponseWriter, r *http.Request) {
	var dbClient = dynamodb.NewFromConfig(cfg)
	var creds adminCreds
	err:= json.NewDecoder(r.Body).Decode(&creds)
	res, err := dbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "admin:" + creds.Username},
			"sk": &types.AttributeValueMemberS{Value: "data"},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal Server error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	if len(res.Item) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid credentials"}
		json.NewEncoder(w).Encode(body)
		return
	}

	var data adminData
	err = attributevalue.UnmarshalMap(res.Item, &data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal Server error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(data.HashedPassword), []byte(creds.Password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid password"}
		json.NewEncoder(w).Encode(body)
		return
	}

	sessionId, err := uuid.NewV7()
	csrf, err := uuid.NewV7()
	currentTime := time.Now().UnixMilli()
	session := localTypes.Session{
		Pk:        "session",
		Sk:        sessionId.String(),
		Role:      "admin",
		Name:      creds.Username,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Status:    localTypes.SessionOngoing,
		CsrfToken: csrf.String(),
	}

	marshalledSession, err := attributevalue.MarshalMap(session)

	_, err = dbClient.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
					Key: map[string]types.AttributeValue{
						"pk": &types.AttributeValueMemberS{Value: "admin:" + creds.Username},
						"sk": &types.AttributeValueMemberS{Value: "data"},
					},
					UpdateExpression: aws.String("SET currentSession=:sessionId, updatedAt=:updatedAt"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":availability": &types.AttributeValueMemberBOOL{Value: false},
						":sessionId":    &types.AttributeValueMemberS{Value: sessionId.String()},
						":name":         &types.AttributeValueMemberS{Value: creds.Username},
						":updatedAt":    &types.AttributeValueMemberN{Value: strconv.FormatInt(currentTime, 10)},
					},
				},
			},
			{
				Put: &types.Put{
					TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
					Item:      marshalledSession,
				},
			},
		},
	})

	userToken := &localTypes.UserTokenType{
		Name: creds.Username,
		Role: "admin",
	}
	encodedToken, err := json.Marshal(userToken)
	hash := sha256.Sum256(encodedToken)
	signature, err := utils.AsymSign(hash[:])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Uable to create session"}
		json.NewEncoder(w).Encode(body)
		return
	}
	userCookie := base64.StdEncoding.EncodeToString(encodedToken) + "," + signature
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Domain:   "localhost",
		Value:    sessionId.String(),
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Domain:   "localhost",
		Value:    csrf.String(),
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_info",
		Domain:   "localhost",
		Value:    userCookie,
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "Login success",
	})
}
