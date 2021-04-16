package http

import (
	"ecommerce/pkg/ecommerce"
	"errors"
	"fmt"
	"net/http"
	"strings"
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

func (h Http) setReqCtxUser(next http.Handler) http.Handler {
	const op = "setReqCtxUser"

	f := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			// no auth header, user is not authenticated
			next.ServeHTTP(w, r)
			return
		}

		bearerTokenSlice := strings.Split(r.Header.Get("Authorization"), " ")
		if len(bearerTokenSlice) != 2 || bearerTokenSlice[0] != "Bearer" {
			h.Response.clientError(w, http.StatusUnauthorized, "wrongly formed authentication header")
			return
		}

		authToken := bearerTokenSlice[1]
		u, err := ecommerce.UserFromAuthToken(authToken)
		if err != nil {
			// user is not logged in
			next.ServeHTTP(w, r)
			return
		}

		ctx := ecommerce.NewUserContext(r.Context(), u)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	return http.HandlerFunc(f)
}

func (h Http) authenticatedOnly(next http.Handler) http.Handler {
	const op = "server.authenticatedOnly"

	f := func(w http.ResponseWriter, r *http.Request) {
		if _, ok := ecommerce.UserFromContext(r.Context()); !ok {
			h.Response.clientError(w, http.StatusUnauthorized, "")
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
