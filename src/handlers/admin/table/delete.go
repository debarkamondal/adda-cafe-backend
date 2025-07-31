package table

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/debarkamondal/adda-cafe-backend/src/clients"
	localTypes "github.com/debarkamondal/adda-cafe-backend/src/types"
	"github.com/golang-jwt/jwt/v5"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	qrToken := r.PathValue("token")
	var tableId string
	parsedToken, err := jwt.ParseWithClaims(qrToken, &localTypes.QRToken{}, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if claims, ok := parsedToken.Claims.(*localTypes.QRToken); ok {
		tableId = claims.TableId
	}
	if tableId == "" {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal server error"}
		json.NewEncoder(w).Encode(body)
		return
	}

	_, err = clients.DBClient.DeleteItem(r.Context(), &dynamodb.DeleteItemInput{
		TableName: aws.String(os.Getenv("DB_TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "table"},
			"sk": &types.AttributeValueMemberS{Value: tableId},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Couldn't create table."}
		json.NewEncoder(w).Encode(body)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
