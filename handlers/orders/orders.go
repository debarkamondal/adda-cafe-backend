package orders

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/debarkamondal/adda-cafe-backend/handlers/ws"
	"github.com/debarkamondal/adda-cafe-backend/types"
	"github.com/google/uuid"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var order types.Order
	json.NewDecoder(r.Body).Decode(&order)

	if len(order.Items) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		body := map[string]any{"message": "Empty item list"}
		json.NewEncoder(w).Encode(body)
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		body := map[string]any{"message": "Internal Server Error"}
		json.NewEncoder(w).Encode(body)
		return
	}
	order.Pk = "order"
	order.Sk = id.String()
	order.CreatedAt = time.Now().UnixMilli()
	ws.Broadcast <- order
}
