package main

import (
	"fmt"
	"net/http"

	adminLogin "github.com/debarkamondal/adda-cafe-backend/handlers/admin/login"
	"github.com/debarkamondal/adda-cafe-backend/handlers/menu"
	"github.com/debarkamondal/adda-cafe-backend/handlers/orders"
	"github.com/debarkamondal/adda-cafe-backend/handlers/reserve"
	"github.com/debarkamondal/adda-cafe-backend/handlers/signin"
	"github.com/debarkamondal/adda-cafe-backend/handlers/user"
	"github.com/debarkamondal/adda-cafe-backend/middlewares"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /signin", signin.Post)
	mux.HandleFunc("POST /admin/login", adminLogin.Post)

	mux.HandleFunc("POST /reserve", reserve.Post)
	mux.HandleFunc("GET /menu", menu.Get)
	mux.HandleFunc("POST /user/admin", user.CreateAdmin)

	// Backend routes
	mux.HandleFunc("POST /menu", menu.Post)
	mux.HandleFunc("DELETE /menu", middlewares.Handle(menu.Delete, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	mux.HandleFunc("PATCH /menu", middlewares.Handle(menu.Patch, []middlewares.Middleware{middlewares.AdminAuthorizer}))

	// mux.HandleFunc("/ws/admin", middlewares.Handle(ws.WsHandler, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	// go ws.HandleBroadcast()

	mux.HandleFunc("POST /orders", middlewares.Handle(orders.Post, []middlewares.Middleware{middlewares.UserAuthorizer}))

	fmt.Println("Listening on port 8080")
	if err := http.ListenAndServe(":8080", http.StripPrefix("/adda",mux)); err != nil {
		fmt.Println(err)
		fmt.Println("Couldn't initiate server on port 8080")
	}
}
