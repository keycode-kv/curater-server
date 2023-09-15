package auth

import (
	"context"
	"crypto/sha256"
	"curater/app"
	"curater/server/api"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const getUserDetailsByID = `select password from users where email = $1`

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authToken struct {
	Token string `json:"token"`
}

type userDetails struct {
	Password string `db:"password"`
}

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

func (s loginRequest) login(ctx context.Context) (token authToken, err error) {
	inputPword := Hash(s.Password)

	var user userDetails

	err = app.GetDB().GetContext(ctx, &user, getUserDetailsByID, s.Email)
	if err != nil {
		return
	}

	if inputPword == user.Password {
		tok := base64.StdEncoding.EncodeToString([]byte(s.Email))
		token = authToken{Token: tok}
	} else {
		err = errors.New("potti")
	}
	return
}

func Hash(input string) string {
	hasher := sha256.New()

	hasher.Write([]byte(input))

	hashBytes := hasher.Sum(nil)

	hashString := hex.EncodeToString(hashBytes)

	return hashString
}
