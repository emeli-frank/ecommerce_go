package http

import (
	"errors"
	"fmt"
	"net/http"
)

func (h Http) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "Close")
				switch t := err.(type) {
				case error:
					h.Response.serverError(w, t)
				default:
					msg := fmt.Sprint("an unknown error:", t)
					h.Response.serverError(w, errors.New(msg))
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
