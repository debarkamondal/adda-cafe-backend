package middlewares

import (
	"fmt"
	"net/http"
	"slices"
)

var allowedOrigins = []string{
	"http://localhost:5173",
}

func CORS(next HandleFunc) HandleFunc {
	fmt.Println("middleware ran")
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Add("Vary", "Origin")
		next(w, r)
	}
}
