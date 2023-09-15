package server

import (
	"curater/auth"
	"curater/newsletter"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func routes(router *mux.Router) {
	router.Handle("/health", health()).Methods(http.MethodGet)

	router.Handle("/login", auth.Login()).Methods(http.MethodPost)

	router.Handle("/newsletter", newsletter.HandleNewsletter()).Methods(http.MethodPost)
}

func health() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Println("health OK", req.Context().Value("user"))
		rw.WriteHeader(200)
	})
}
