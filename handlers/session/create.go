package session

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type tableToken struct {
	Id string `json:"id"`
	jwt.RegisteredClaims
}

func Create(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	tempToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"id": "123",
	})
	_, err := tempToken.SignedString([]byte("test"))

	parsedToken, err := jwt.ParseWithClaims(token, &tableToken{}, func(t *jwt.Token) (any, error) {
		return []byte("test"), nil
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Invalid Token"}
		json.NewEncoder(w).Encode(body)
		return
	}
	if tempClaim, ok := parsedToken.Claims.(*tableToken); ok {
		fmt.Println(tempClaim.Id)
	}

}
