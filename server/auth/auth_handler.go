package auth

import (
	"curater/server/api"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func Login() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		ctx := req.Context()

		var loginReq loginRequest
		err := json.NewDecoder(req.Body).Decode(&loginReq)
		if err != nil {
			fmt.Println("error while unmarshalling login request")
			api.RespondWithJSON(rw, http.StatusBadRequest, []byte(""))
			return
		}

		auth, err := loginReq.login(ctx)
		if err != nil {
			fmt.Println("auth potti", err.Error())
			api.RespondWithJSON(rw, http.StatusUnauthorized, []byte(""))
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, auth)
	})
}

func SignUp() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		ctx := req.Context()

		var signupReq signupRequest
		err := json.NewDecoder(req.Body).Decode(&signupReq)
		if err != nil {
			fmt.Println("error while unmarshalling signup request")
			api.RespondWithJSON(rw, http.StatusBadRequest, []byte(""))
			return
		}

		auth, err := signupReq.signup(ctx)
		if err != nil {
			fmt.Println("signup potti", err.Error())
			api.RespondWithJSON(rw, http.StatusBadRequest, err.Error())
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, auth)
	})
}

func GetProfile() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		ctx := req.Context()

		userInfo, err := getProfile(ctx)
		if err != nil {
			fmt.Println("error getting user profile: ", err.Error())
			if err == sql.ErrNoRows {
				api.RespondWithJSON(rw, http.StatusNotFound, "user not found")
			}
			api.RespondWithJSON(rw, http.StatusInternalServerError, err.Error())
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, userInfo)
	})
}

func GetUserCollections() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		collections, err := getCollections(ctx)
		if err != nil {
			fmt.Println("error getting user collections: ", err.Error())
			api.RespondWithJSON(rw, http.StatusInternalServerError, err.Error())
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, collections)
	})
}

func GetRedirectConfig() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		ctx := req.Context()

		userInfo, err := getuserConfig(ctx)
		if err != nil {
			fmt.Println("error getting user redirect config: ", err.Error())
			if err == sql.ErrNoRows {
				api.RespondWithJSON(rw, http.StatusNotFound, "user not found")
			}
			api.RespondWithJSON(rw, http.StatusInternalServerError, err.Error())
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, userInfo)
	})
}
