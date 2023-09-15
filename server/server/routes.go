package server

import (
	"curater/auth"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func routes(router *mux.Router) {
	router.Handle("/health", health()).Methods(http.MethodGet)

	router.Handle("/login", auth.Login()).Methods(http.MethodPost)
	router.Handle("/signup", auth.SignUp()).Methods(http.MethodPost)
	router.Handle("/profile", auth.GetProfile()).Methods(http.MethodGet)
	router.Handle("/users/{id}/collections", auth.GetUserCollections()).Methods(http.MethodGet)
	router.Handle("/configuration", auth.GetRedirectConfig()).Methods(http.MethodGet)
}

func health() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Println("health OK", req.Context().Value("user"))
		rw.WriteHeader(200)
	})
}
