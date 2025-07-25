package login

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	localTypes "github.com/debarkamondal/adda-cafe-backend/src/types"

	"github.com/debarkamondal/adda-cafe-backend/src/utils"
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
	err := json.NewDecoder(r.Body).Decode(&creds)
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
	session := localTypes.BackendSession{
		Pk:        "session:backend",
		Sk:        sessionId.String(),
		Role:      "admin",
		Name:      creds.Username,
		CreatedAt: currentTime,
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
						":sessionId": &types.AttributeValueMemberS{Value: sessionId.String()},
						":updatedAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(currentTime, 10)},
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
		body := map[string]any{"message": "Internal Server Error"}
		json.NewEncoder(w).Encode(body)
		return
	}

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

	// stripping subdomain
	domainArr := strings.SplitN(os.Getenv("BACKEND_DOMAIN"), ".", 2)
	var rootDomain string
	// setting rootDomain to localhost or baseDomain
	if len(domainArr) > 1 {
		rootDomain = "." + domainArr[1]
	} else {
		rootDomain = os.Getenv("BACKEND_DOMAIN")
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Domain:   os.Getenv("BACKEND_DOMAIN"),
		Path:     os.Getenv("URI_PREFIX"),
		Value:    sessionId.String(),
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "csrf_token",
		Domain: rootDomain,
		Path:   "/",
		// Path:     os.Getenv("URI_PREFIX"),
		Value:    csrf.String(),
		MaxAge:   10800,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_info",
		Domain:   rootDomain,
		Path:     "/",
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
