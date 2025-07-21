package main

import (
	"fmt"
	"net/http"
	"os"

	adminLogin "github.com/debarkamondal/adda-cafe-backend/handlers/admin/login"
	"github.com/debarkamondal/adda-cafe-backend/handlers/admin/ws"
	"github.com/debarkamondal/adda-cafe-backend/handlers/menu"
	"github.com/debarkamondal/adda-cafe-backend/handlers/orders"
	"github.com/debarkamondal/adda-cafe-backend/handlers/reserve"
	"github.com/debarkamondal/adda-cafe-backend/handlers/signin"
	"github.com/debarkamondal/adda-cafe-backend/handlers/user"
	"github.com/debarkamondal/adda-cafe-backend/middlewares"
)

func main() {
	mux := http.NewServeMux()

	// Public routes (Unauthenticated)
	mux.HandleFunc("POST /signin", signin.Post)
	mux.HandleFunc("POST /admin/login", adminLogin.Post)

	mux.HandleFunc("POST /reserve", reserve.Post)
	mux.HandleFunc("GET /menu", menu.Get)

	// Backend routes (Authenticated)
	mux.HandleFunc("POST /user/admin", user.CreateAdmin)

	mux.HandleFunc("POST /admin/menu", middlewares.Handle(menu.Post, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	mux.HandleFunc("DELETE /admin/menu", middlewares.Handle(menu.Delete, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	mux.HandleFunc("PATCH /admin/menu", middlewares.Handle(menu.Patch, []middlewares.Middleware{middlewares.AdminAuthorizer}))

	// Authentication and authorization for the websocket is being done by the handler itself
	mux.HandleFunc("/admin/ws", ws.WsHandler)
	go ws.HandleBroadcast()

	// User routes (Authenticated)
	mux.HandleFunc("POST /orders", middlewares.Handle(orders.Post, []middlewares.Middleware{middlewares.UserAuthorizer}))

	if os.Getenv("PROXY") == "true" {
		fmt.Println("Listening on port 8080 (proxied)")
		if err := http.ListenAndServe(":8080", http.StripPrefix(os.Getenv("URI_PREFIX"), mux)); err != nil {
			fmt.Println(err)
			fmt.Println("Couldn't initiate server on port 8080")
		}
	} else {
		fmt.Println("Listening on port 8080 (unproxied)")
		if err := http.ListenAndServe(":8080", http.StripPrefix(os.Getenv("URI_PREFIX"), middlewares.CORS(mux))); err != nil {
			fmt.Println(err)
			fmt.Println("Couldn't initiate server on port 8080")
		}
	}

}
