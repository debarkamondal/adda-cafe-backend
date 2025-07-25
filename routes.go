package main

import (
	"net/http"

	adminLogin "github.com/debarkamondal/adda-cafe-backend/handlers/admin/login"
	"github.com/debarkamondal/adda-cafe-backend/handlers/admin/pending"
	"github.com/debarkamondal/adda-cafe-backend/handlers/admin/ws"
	"github.com/debarkamondal/adda-cafe-backend/handlers/menu"
	"github.com/debarkamondal/adda-cafe-backend/handlers/orders"
	"github.com/debarkamondal/adda-cafe-backend/handlers/reserve"
	"github.com/debarkamondal/adda-cafe-backend/handlers/signin"
	"github.com/debarkamondal/adda-cafe-backend/handlers/user"
	"github.com/debarkamondal/adda-cafe-backend/middlewares"
)

func InitRoutes(mux *http.ServeMux) *http.ServeMux {
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

	mux.HandleFunc("DELETE /admin/pending", middlewares.Handle(pending.Delete, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	mux.HandleFunc("PUT /admin/pending", middlewares.Handle(pending.Put, []middlewares.Middleware{middlewares.AdminAuthorizer}))

	mux.HandleFunc("/admin/ws", ws.WsHandler) // Authentication and authorization for the websocket is being done by the handler itself

	// User routes (Authenticated)
	mux.HandleFunc("POST /orders", middlewares.Handle(orders.Post, []middlewares.Middleware{middlewares.UserAuthorizer}))
	return mux
}
