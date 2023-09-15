package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func RespondWithJSON(rw http.ResponseWriter, status int, response interface{}) {
	respBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("marshal cheyyan pattanilla")
		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(respBytes)
}
