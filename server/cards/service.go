package cards

import (
	"context"
	"curater/server/api"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Filter struct {
	Type       string
	Search     string
	Collection string
	Tags       []string
}

type Cards struct {
	Cards []Card `json:"cards"`
}

type tagListRespose struct {
	Tags []string `json:"tags"`
}

type updateCardRequest struct {
	ID           int64  `json:"id"`
	Status       string `json:"status"`
	CollectionID int64  `json:"collection_id"`
	IsViewed     bool   `json:"is_viewed"`
}

type updateCardResponse struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}
type Comments struct {
	Comments []Comment `json:"comments"`
}

type commentRequest struct {
	ContentID int64  `json:"content_id"`
	Comment   string `json:"comment"`
	UserID    int64  `json:"user_id"`
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
		collection := query.Get("collection")
		tags := query["tags"]

		filters := Filter{
			Type:       cardType,
			Search:     search,
			Collection: collection,
			Tags:       tags,
		}

		userID := req.Context().Value("user")
		resp, err := GetCardsForUser(userID.(string), filters)
		if err != nil {
			fmt.Print("error getting cards list, error: ", err.Error())
			api.RespondWithJSON(rw, http.StatusInternalServerError, "error getting cards list")
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, resp)
	})
}

func GetCardByID() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		vars := mux.Vars(req)
		cardID, _ := strconv.Atoi(vars["id"])
		userEmail := req.Context().Value("user")
		resp, err := GetCardByIDForUser(req.Context(), userEmail.(string), int64(cardID))
		if err != nil {
			fmt.Print("error getting card details, error: ", err.Error())
			api.RespondWithJSON(rw, http.StatusInternalServerError, "error getting card details")
			return
		}
		api.RespondWithJSON(rw, http.StatusOK, resp)
	})
}

func getAllTags(ctx context.Context) (tags tagListRespose, err error) {

	tagList, err := getTagList(ctx)
	if err != nil {
		fmt.Printf("error getting tag list, error: %s\n", err.Error())
		return
	}
	tags.Tags = tagList
	return
}

func updateCard(ctx context.Context, request updateCardRequest) (response updateCardResponse, err error) {
	if request.CollectionID == 0 && request.Status == "saved" {
		userInfo, err := getUserByEmail(ctx)
		if err != nil {
			fmt.Printf("error getting user: %s, error: %s\n", ctx.Value("user"), err.Error())
			return response, err
		}
		collectionID, err := getDefaultCollectionIDByUserID(ctx, userInfo.ID)
		if err != nil {
			fmt.Printf("error getting user collection for user_id: %d, error: %s\n", userInfo.ID, err.Error())
			return response, err
		}
		request.CollectionID = collectionID
	}
	card, err := updateCardInfo(ctx, request)
	if err != nil {
		fmt.Printf("error updating card details, request: %+v, error: %s", request, err.Error())
		return
	}
	response = updateCardResponse{
		ID:     int64(card.ID),
		Status: card.Status,
	}
	return
}

func GetCommentsByID() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		vars := mux.Vars(req)
		contentID := vars["id"]
		userID := req.Context().Value("user")
		resp, err := getComments(req.Context(), userID.(string), contentID)
		if err != nil {
			fmt.Print("error getting comments, error: ", err.Error())
			api.RespondWithJSON(rw, http.StatusInternalServerError, "error getting comments")
			return
		}

		api.RespondWithJSON(rw, http.StatusOK, resp)

	})
}

func getComments(ctx context.Context, userID string, contentID string) (Comments, error) {

	comments := Comments{
		Comments: []Comment{},
	}
	commentList, err := getCommentsByContentID(userID, contentID)
	if err != nil {
		fmt.Printf("error getting article count for user %s, error: %s\n", ctx.Value("user"), err.Error())
		return comments, err
	}

	for _, item := range commentList {
		comments.Comments = append(comments.Comments, Comment{
			ID:          item.ID,
			Comment:     item.Comment,
			User:        item.User,
			CommentedAt: item.CommentedAt,
		})
	}

	return comments, err
}

func postComment(ctx context.Context, request commentRequest) (response Comment, err error) {
	userInfo, err := getUserByEmail(ctx)
	if err != nil {
		fmt.Printf("error getting user: %s, error: %s\n", ctx.Value("user"), err.Error())
		return
	}

	request.UserID = userInfo.ID
	response, err = createComment(ctx, request)
	if err != nil {
		fmt.Printf("error updating card details, request: %+v, error: %s", request, err.Error())
		return
	}
	return
}

func PostRating() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var request PostRatingRequest

		vars := mux.Vars(req)
		contentID, _ := strconv.Atoi(vars["id"])
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			api.RespondWithJSON(rw, http.StatusBadRequest, "error decoding request")
			return
		}
		request.ContentID = contentID
		userID := req.Context().Value("user")
		err = InsertRating(userID.(string), request)
		if err != nil {
			fmt.Print("error adding rating, error: ", err.Error())
			api.RespondWithJSON(rw, http.StatusInternalServerError, "error adding rating")
			return
		}
		api.RespondWithJSON(rw, http.StatusOK, "rating has been submitted")

	})
}
