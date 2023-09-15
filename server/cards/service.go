package cards

import (
	"curater/server/api"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type Filter struct {
	Type   string
	Search string
	Tags   []string
}

type Cards struct {
	Cards []Card `json:"cards"`
}

type Comments struct {
	Comments []Comment `json:"comments,omitempty"`
}

type PostRatingRequest struct {
	Rating    int `json:"rating"`
	ContentID int `json:"content_id"`
}

func GetCards() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		query := req.URL.Query()
		cardType := query.Get("type")
		search := query.Get("search")
		tags := query["tags"]

		filters := Filter{
			Type:   cardType,
			Search: search,
			Tags:   tags,
		}

		var cards Cards

		userID := req.Context().Value("user")
		resp, err := GetCardsForUser(userID.(string), filters)
		if err != nil {
			fmt.Print("errorr pottii", err.Error())
		}

		cards.Cards = resp

		api.RespondWithJSON(rw, http.StatusOK, cards)
	})
}

func GetCardByID() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		vars := mux.Vars(req)
		cardID := vars["id"]
		userID := req.Context().Value("user")
		resp, err := GetCardByIDForUser(userID.(string), cardID)
		if err != nil {
			fmt.Print("errorr pottii", err.Error())
		}
		api.RespondWithJSON(rw, http.StatusOK, resp)

	})
}

func GetCommentsByID() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		var comments Comments
		vars := mux.Vars(req)
		contentID := vars["id"]
		userID := req.Context().Value("user")
		resp, err := GetCommentsByContentID(userID.(string), contentID)
		if err != nil {
			fmt.Print("errorr pottii", err.Error())
		}

		comments.Comments = resp
		api.RespondWithJSON(rw, http.StatusOK, comments)

	})
}

func PostRating() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var request PostRatingRequest

		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			api.RespondWithJSON(rw, http.StatusBadRequest, "error decoding request")
			return
		}
		userID := req.Context().Value("user")
		err = InsertRating(userID.(string), request)
		if err != nil {
			fmt.Print("errorr pottii", err.Error())
			api.RespondWithJSON(rw, http.StatusBadRequest, err.Error())
			return
		}
		api.RespondWithJSON(rw, http.StatusOK, "comment has been posted")

	})
}
