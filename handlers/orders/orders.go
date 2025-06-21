package orders

import (
	"encoding/json"
	"net/http"

	"github.com/debarkamondal/adda-cafe-backend/handlers/ws"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)
	ws.Broadcast <- body
}
