package main

import (
	"fmt"
	"net/http"

	"github.com/debarkamondal/adda-cafe-backend/handlers/menu"
	"github.com/debarkamondal/adda-cafe-backend/handlers/orders"
	"github.com/debarkamondal/adda-cafe-backend/handlers/reserve"
	"github.com/debarkamondal/adda-cafe-backend/handlers/session"
	"github.com/debarkamondal/adda-cafe-backend/handlers/ws"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /session", session.Create)

	mux.HandleFunc("GET /menu", menu.Get)
	mux.HandleFunc("POST /menu", menu.Create)
	mux.HandleFunc("PATCH /menu/{id}", menu.Update)
	mux.HandleFunc("DELETE /menu", menu.Delete)

	mux.HandleFunc("/ws/admin", ws.WsHandler)
	go ws.HandleBroadcast()

	mux.HandleFunc("POST /orders", orders.Create)
	mux.HandleFunc("GET /reserve", reserve.Create)

	fmt.Println("Listening on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("Couldn't initiate server on port 8080")
	}
}
