package reserve

import (
	"context"
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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
var dbClient = dynamodb.NewFromConfig(cfg)

func Post(w http.ResponseWriter, r *http.Request) {

	token := r.URL.Query().Get("token")
	var tableId string

	parsedToken, err := jwt.ParseWithClaims(token, &localTypes.TableToken{}, func(t *jwt.Token) (any, error) {
		return []byte("test"), nil
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid Token"}
		json.NewEncoder(w).Encode(body)
		return
	}

	if claims, ok := parsedToken.Claims.(*localTypes.TableToken); ok {
		tableId = claims.Id
	}
	if tableId == "" {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid Token"}
		json.NewEncoder(w).Encode(body)
		return
	}

	res, err := dbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "table"},
			"sk": &types.AttributeValueMemberS{Value: tableId},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "DB error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	if res.Item == nil {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid table"}
		json.NewEncoder(w).Encode(body)
		return
	}

	var table localTypes.Table
	attributevalue.UnmarshalMap(res.Item, &table)
	if !table.IsAvailable {
		w.WriteHeader(http.StatusConflict)
		body := map[string]any{"message": "Table is already reserved. If free please contact us."}
		json.NewEncoder(w).Encode(body)
		return
	}

	uid, err := uuid.NewV7()
	csrf, err := uuid.NewV7()
	currentTime := time.Now().UnixMilli()
	session := localTypes.Session{
		Pk:        "session",
		Sk:        uid.String(),
		TableId:   tableId,
		CreatedAt: currentTime,
		Orders:    []string{},
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
						"pk": &types.AttributeValueMemberS{Value: "table"}, // Partition Key
						"sk": &types.AttributeValueMemberS{Value: tableId}, // Sort Key
					},
					UpdateExpression: aws.String("SET isAvailable = :availability, currentSession=:sessionId, updatedAt=:updatedAt"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":availability": &types.AttributeValueMemberBOOL{Value: false},
						":sessionId":    &types.AttributeValueMemberS{Value: uid.String()},
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

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Unable to reserve the table. Please contact us."}
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
	body := map[string]any{"message": "table reserved"}
	json.NewEncoder(w).Encode(body)
}
