package main

import (
	"net/http"

	adminLogin "github.com/debarkamondal/adda-cafe-backend/src/handlers/admin/login"
	adminMenu "github.com/debarkamondal/adda-cafe-backend/src/handlers/admin/menu"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/admin/pending"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/admin/table"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/admin/ws"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/menu"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/orders"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/reserve"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/signin"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/user"
	"github.com/debarkamondal/adda-cafe-backend/src/middlewares"
)

func InitRoutes(mux *http.ServeMux) *http.ServeMux {
	// Public routes (Unauthenticated)
	mux.HandleFunc("POST /signin", signin.Post)
	mux.HandleFunc("POST /admin/login", adminLogin.Post)

	mux.HandleFunc("POST /reserve", reserve.Post)

	mux.HandleFunc("GET /menu", menu.Get)
	mux.HandleFunc("GET /menu/{id}", menu.GetById)

	// Backend routes (Authenticated)
	mux.HandleFunc("POST /user/admin", user.CreateAdmin)

	mux.HandleFunc("POST /admin/menu", middlewares.Handle(adminMenu.Post, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	mux.HandleFunc("DELETE /admin/menu/{id}", middlewares.Handle(adminMenu.Delete, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	mux.HandleFunc("PATCH /admin/menu", middlewares.Handle(adminMenu.Patch, []middlewares.Middleware{middlewares.AdminAuthorizer}))

	mux.HandleFunc("DELETE /admin/pending", middlewares.Handle(pending.Delete, []middlewares.Middleware{middlewares.AdminAuthorizer}))
	mux.HandleFunc("PUT /admin/pending", middlewares.Handle(pending.Put, []middlewares.Middleware{middlewares.AdminAuthorizer}))

	mux.HandleFunc("POST /admin/table/{name}", middlewares.Handle(table.Post, []middlewares.Middleware{middlewares.AdminAuthorizer}))

	mux.HandleFunc("/admin/ws", ws.WsHandler) // Authentication and authorization for the websocket is being done by the handler itself

	// User routes (Authenticated)
	mux.HandleFunc("POST /order", middlewares.Handle(orders.Post, []middlewares.Middleware{middlewares.UserAuthorizer}))
	return mux
}
