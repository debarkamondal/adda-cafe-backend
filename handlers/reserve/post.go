package reserve

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	// "os"
	// "strconv"
	// "time"

	// "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/debarkamondal/adda-cafe-backend/types"
	"github.com/debarkamondal/adda-cafe-backend/utils"

	// "github.com/golang-jwt/jwt/v5"

	// "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	// "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	// "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	// localTypes "github.com/debarkamondal/adda-cafe-backend/types"
	// "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

func Post(w http.ResponseWriter, r *http.Request) {
	// var dbClient = dynamodb.NewFromConfig(cfg)

	var body struct {
		Name   string `json:"name" dynamodbav:"name"`
		Token  string `json:"token" dynamodbav:"token"`
		Phone  int64  `json:"phone" dynamodbav:"phone"`
		Coords struct {
			Lat  float64 `json:"lat" dynamodbav:"lat"`
			Long float64 `json:"long" dynamodbav:"long"`
		} `json:"coords" dynamodbav:"coords"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	// var tableId string
	//
	// parsedToken, err := jwt.ParseWithClaims(body.Token, &localTypes.TableToken{}, func(t *jwt.Token) (any, error) {
	// 	return []byte("test"), nil
	// })
	//
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	body := map[string]any{"message": "Invalid Token"}
	// 	json.NewEncoder(w).Encode(body)
	// 	return
	// }
	//
	// if claims, ok := parsedToken.Claims.(*localTypes.TableToken); ok {
	// 	tableId = claims.Id
	// }
	// if tableId == "" {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	body := map[string]any{"message": "Invalid Token"}
	// 	json.NewEncoder(w).Encode(body)
	// 	return
	// }
	//
	// res, err := dbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
	// 	TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
	// 	Key: map[string]types.AttributeValue{
	// 		"pk": &types.AttributeValueMemberS{Value: "table"},
	// 		"sk": &types.AttributeValueMemberS{Value: tableId},
	// 	},
	// })
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	body := map[string]any{"message": "DB error"}
	// 	json.NewEncoder(w).Encode(body)
	// 	return
	// }
	// if res.Item == nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	body := map[string]any{"message": "Invalid table"}
	// 	json.NewEncoder(w).Encode(body)
	// 	return
	// }
	//
	// var table localTypes.Table
	// attributevalue.UnmarshalMap(res.Item, &table)
	// if !table.IsAvailable {
	// 	w.WriteHeader(http.StatusConflict)
	// 	body := map[string]any{"message": "Table is already reserved. Please contact us."}
	// 	json.NewEncoder(w).Encode(body)
	// 	return
	// }

	uid, err := uuid.NewV7()
	csrf, err := uuid.NewV7()
	// currentTime := time.Now().UnixMilli()
	// session := localTypes.Session{
	// 	Pk:        "session",
	// 	Sk:        uid.String(),
	// Role:	"user",
	// 	TableId:   tableId,
	// 	Name:      body.Name,
	// 	Phone:     body.Phone,
	// 	CreatedAt: currentTime,
	// 	Orders:    []string{},
	// 	UpdatedAt: currentTime,
	// 	Status:    localTypes.SessionOngoing,
	// 	CsrfToken: csrf.String(),
	// }
	//
	// marshalledSession, err := attributevalue.MarshalMap(session)
	//
	// _, err = dbClient.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
	// 	TransactItems: []types.TransactWriteItem{
	// 		{
	// 			Update: &types.Update{
	// 				TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
	// 				Key: map[string]types.AttributeValue{
	// 					"pk": &types.AttributeValueMemberS{Value: "table"}, // Partition Key
	// 					"sk": &types.AttributeValueMemberS{Value: tableId}, // Sort Key
	// 				},
	// 				UpdateExpression: aws.String("SET isAvailable = :availability, currentSession=:sessionId, updatedAt=:updatedAt, name=:name, phone=:phone"),
	// 				ExpressionAttributeValues: map[string]types.AttributeValue{
	// 					":availability": &types.AttributeValueMemberBOOL{Value: false},
	// 					":sessionId":    &types.AttributeValueMemberS{Value: uid.String()},
	// 					":name":         &types.AttributeValueMemberS{Value: body.Name},
	// 					":updatedAt":    &types.AttributeValueMemberN{Value: strconv.FormatInt(currentTime, 10)},
	// 					":phone":        &types.AttributeValueMemberN{Value: strconv.FormatInt(body.Phone, 10)},
	// 				},
	// 			},
	// 		},
	// 		{
	// 			Put: &types.Put{
	// 				TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
	// 				Item:      marshalledSession,
	// 			},
	// 		},
	// 	},
	// })
	//

	userToken := &types.UserTokenType{
		Name: body.Name,
		Role: "user",
	}
	encodedToken, err := json.Marshal(userToken)
	hash := sha256.Sum256(encodedToken)
	signature, err := utils.AsymSign(hash[:])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Unable to reserve the table. Please contact us."}
		json.NewEncoder(w).Encode(body)
		return
	}
	userCookie := base64.StdEncoding.EncodeToString(encodedToken) + "," + signature

	// removing subdomain from the url for the frontend cookie
	domainArr := strings.SplitN(os.Getenv("BACKEND_DOMAIN"), ".", 2)
	var domain string
	if len(domainArr) > 1 {
		domain = "." + domainArr[1]
	} else {
		domain = os.Getenv("BACKEND_DOMAIN")
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Domain:   os.Getenv("BACKEND_DOMAIN"),
		Path:     os.Getenv("URI_PREFIX"),
		Value:    uid.String(),
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Domain:   os.Getenv("BACKEND_DOMAIN"),
		Path:     os.Getenv("URI_PREFIX"),
		Value:    csrf.String(),
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_info",
		Domain:   domain,
		Path:     "/",
		Value:    userCookie,
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "table reserved",
	})
}
