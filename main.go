package main

import (
	"fmt"
	"net/http"

	"github.com/debarkamondal/adda-cafe-backend/handlers/menu"
	"github.com/debarkamondal/adda-cafe-backend/handlers/orders"
	"github.com/debarkamondal/adda-cafe-backend/handlers/reserve"
	"github.com/debarkamondal/adda-cafe-backend/handlers/signin"
	"github.com/debarkamondal/adda-cafe-backend/handlers/ws"
	"github.com/debarkamondal/adda-cafe-backend/middlewares"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /signin", signin.Create)

	mux.HandleFunc("GET /reserve", reserve.Create)
	mux.HandleFunc("GET /menu", menu.Get)

	// Backend routes
	mux.HandleFunc("POST /menu", middlewares.Handle(menu.Create, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	mux.HandleFunc("DELETE /menu", middlewares.Handle(menu.Delete, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	mux.HandleFunc("PATCH /menu", middlewares.Handle(menu.Update, []middlewares.Middleware{middlewares.AdminAuthorizer}))

	mux.HandleFunc("/ws/admin", middlewares.Handle(ws.WsHandler, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	go ws.HandleBroadcast()

	mux.HandleFunc("POST /orders", middlewares.Handle(orders.Create, []middlewares.Middleware{middlewares.UserAuthorizer}))

	fmt.Println("Listening on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't initiate server on port 8080")
	}
}
