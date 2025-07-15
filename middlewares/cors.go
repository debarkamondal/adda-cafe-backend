package middlewares

import (
	"net/http"
	"os"
	"slices"
)

var allowedOrigins = []string{
	"http://localhost:5173",
	os.Getenv("BACKEND_URL"),
}

func CORS(mux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		origin := r.Header.Get("Origin")
		if slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Add("Vary", "Origin")
		mux.ServeHTTP(w, r)
	})
}
