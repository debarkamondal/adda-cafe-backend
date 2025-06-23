package middlewares

import (
	"net/http"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)
type Middleware func(HandleFunc) HandleFunc

func Handle(final HandleFunc, middlewares []Middleware) HandleFunc {
	if final == nil {
		panic("no final handler")
	}
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final)
	}
	return final
}
