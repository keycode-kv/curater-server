package server

import (
	"context"
	"curater/server/api"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func startHTTPServer() (err error) {
	router := mux.NewRouter()

	router.Use(loginMiddleware)

	routes(router)
	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	corsHandler := handlers.CORS(headersOk, originsOk, methodsOk)(router)

	port := 8082
	fmt.Printf("Server is running on port %d...\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), corsHandler)
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
			"/newsletter",
		}
		for _, val := range publicEndpoints {
			if val == r.RequestURI {
				next.ServeHTTP(w, r)
				return
			}
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			fmt.Println("Authorization header is missing")
			return
		}

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
