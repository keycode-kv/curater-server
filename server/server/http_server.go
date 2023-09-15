package server

import (
	"context"
	"curater/server/api"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func startHTTPServer() (err error) {
	router := mux.NewRouter()

	router.Use(loginMiddleware)

	routes(router)

	port := 8082
	fmt.Printf("Server is running on port %d...\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	if err != nil {
		fmt.Println(err)
	}

	return
}

func loginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicEndpoints := []string{
			"/login",
			"/signup",
		}
		for _, val := range publicEndpoints {
			if val == r.RequestURI {
				next.ServeHTTP(w, r)
				return
			}
		}

		token := r.Header.Get("Authorization")
		user, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			fmt.Println("error decoding token ", err.Error())
			api.RespondWithJSON(w, http.StatusUnauthorized, "")
			return
		}
		ctx := context.WithValue(r.Context(), "user", string(user))

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
