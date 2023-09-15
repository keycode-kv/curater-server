package server

import (
	"curater/auth"
	"curater/cards"
	"curater/newsletter"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func routes(router *mux.Router) {
	router.Handle("/health", health()).Methods(http.MethodGet)

	router.Handle("/login", auth.Login()).Methods(http.MethodPost)

	router.Handle("/newsletter", newsletter.HandleNewsletter()).Methods(http.MethodPost)
	router.Handle("/signup", auth.SignUp()).Methods(http.MethodPost)

	router.Handle("/profile", auth.GetProfile()).Methods(http.MethodGet)
	router.Handle("/collections", auth.GetUserCollections()).Methods(http.MethodGet)
	router.Handle("/configuration", auth.GetRedirectConfig()).Methods(http.MethodGet)

	router.Handle("/cards", cards.GetCards()).Methods(http.MethodGet)
	router.Handle("/cards/{id}", cards.GetCardByID()).Methods(http.MethodGet)
	router.Handle("/cards/{id}", cards.UpdateCard()).Methods(http.MethodPut)
	router.Handle("/tags", cards.GetTags()).Methods(http.MethodGet)

	router.Handle("/contents/{id}/comments", cards.PostComment()).Methods(http.MethodPost)
	router.Handle("/contents/{id}/comments", cards.GetCommentsByID()).Methods(http.MethodGet)
	router.Handle("/rating", cards.PostRating()).Methods(http.MethodPost)
}

func health() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Println("health OK", req.Context().Value("user"))
		rw.WriteHeader(200)
	})
}
