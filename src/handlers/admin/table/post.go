package table

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/debarkamondal/adda-cafe-backend/src/clients"
	localTypes "github.com/debarkamondal/adda-cafe-backend/src/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func Post(w http.ResponseWriter, r *http.Request) {
	tableName := r.PathValue("name")
	tableId, err := uuid.NewV7()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal server error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	currentTime := time.Now().UnixMilli()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal server error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	qrClaims := localTypes.QRToken{
		TableId: tableId.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	qrToken := jwt.NewWithClaims(jwt.SigningMethodHS256, qrClaims)
	signedToken, err := qrToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal server error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	tableData := localTypes.Table{
		Pk:          "table",
		Sk:          tableId.String(),
		QRToken:     signedToken,
		Title:       tableName,
		IsAvailable: true,
		UpdatedAt:   currentTime,
	}
	marshalledTable, err := attributevalue.MarshalMap(tableData)
	_, err = clients.DBClient.PutItem(r.Context(), &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Item:      marshalledTable,
	})
	// _, err = clients.DBClient.TransactWriteItems(r.Context(), &dynamodb.TransactWriteItemsInput{
	// 	TransactItems: []types.TransactWriteItem{
	// 		{
	// 			Put: &types.Put{
	// 				TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
	// 				Item:      marshalledTable,
	// 			},
	// 		},
	// 		{
	// 			Update: &types.Update{
	// 				TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
	// 				Key: map[string]types.AttributeValue{
	// 					"pk": &types.AttributeValueMemberS{Value: "qrTokens"},
	// 					"sk": &types.AttributeValueMemberS{Value: "tables"},
	// 				},
	// 				UpdateExpression: aws.String("SET #tableId = :qrToken "),
	// 				ExpressionAttributeNames: map[string]string{
	// 					"#tableId": tableId.String(),
	// 				},
	// 				ExpressionAttributeValues: map[string]types.AttributeValue{
	// 					":qrToken": &types.AttributeValueMemberS{Value: signedToken},
	// 				},
	// 			},
	// 		},
	// 	},
	// })
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Couldn't create table."}
		json.NewEncoder(w).Encode(body)
		return
	}
	w.WriteHeader(http.StatusOK)
	body := map[string]any{"name": tableName,
		"token": signedToken}
	json.NewEncoder(w).Encode(body)
}
