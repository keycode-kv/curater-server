package cards

import (
	"curater/server/api"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func GetTags() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		tags, err := getAllTags(req.Context())
		if err != nil {
			fmt.Println("error getting user collections: ", err.Error())
			api.RespondWithJSON(rw, http.StatusInternalServerError, err.Error())
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, tags)
	})
}

func UpdateCard() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		vars := mux.Vars(req)
		cardID, _ := strconv.Atoi(vars["id"])

		var request updateCardRequest
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			fmt.Println("error while unmarshalling update card request")
			api.RespondWithJSON(rw, http.StatusBadRequest, []byte(""))
			return
		}
		request.ID = int64(cardID)
		cardInfo, err := updateCard(ctx, request)
		if err != nil {
			fmt.Println("error updating card details, error: ", err.Error())
			api.RespondWithJSON(rw, http.StatusInternalServerError, err.Error())
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, cardInfo)
	})
}

func PostComment() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		vars := mux.Vars(req)
		contentID, _ := strconv.Atoi(vars["id"])

		var request commentRequest
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			fmt.Println("error while unmarshalling post comment request")
			api.RespondWithJSON(rw, http.StatusBadRequest, []byte(""))
			return
		}
		request.ContentID = int64(contentID)
		commentInfo, err := postComment(ctx, request)
		if err != nil {
			fmt.Println("error updating card details, error: ", err.Error())
			api.RespondWithJSON(rw, http.StatusInternalServerError, err.Error())
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, commentInfo)
	})
}
