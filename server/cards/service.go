package cards

import (
	"curater/server/api"
	"fmt"
	"net/http"
)

type Filter struct {
	Type   string
	Search string
	Tags   []string
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

		userID := req.Context().Value("user")
		resp, err := GetCardsForUser(userID.(string), filters)
		if err != nil {
			fmt.Print("errorr pottii", err.Error())
		}

		api.RespondWithJSON(rw, http.StatusOK, resp)
	})
}
