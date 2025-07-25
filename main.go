package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/debarkamondal/adda-cafe-backend/src/clients"
	"github.com/debarkamondal/adda-cafe-backend/src/handlers/admin/ws"
	"github.com/debarkamondal/adda-cafe-backend/src/middlewares"
)

func main() {
	clients.Init()

	mux := InitRoutes(http.NewServeMux())

	// Handels websocket broadcast
	go ws.HandleBroadcast()

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
