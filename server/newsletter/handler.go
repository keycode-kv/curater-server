package newsletter

import (
	"context"
	"curater/app"
	"curater/server/api"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gojektech/heimdall/v6/httpclient"
)

const setContent = `INSERT INTO public."content"
("content", title, source_email,plain_text,read_time,  created_at, updated_at)
VALUES($1,$2,$3,$4,$5,now(), now())
RETURNING id;`

const getIDBySubjectAndSender = `select id from "content" where title = $1 and source_email = $2`

const getUserIDFromEmail = `select id from "users" where redirect_email = $1`

const createCard = `insert into cards
("user_id", "content_id", "status", "is_viewed", "created_at", "updated_at")
values($1, $2, $3, false, now(), now())
returning id;`

type Headers struct {
	UserEmail string `json:"to"`
	Subject   string `json:"subject"`
}

type NewsLetter struct {
	Header      Headers `json:"headers"`
	PlainText   string  `json:"plain"`
	HTMLContent string  `json:"html"`
}

type SummaryRequest struct {
	ID int64 `json:"id"`
}

func HandleNewsletter() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		var newsletter NewsLetter

		err := json.NewDecoder(req.Body).Decode(&newsletter)
		if err != nil {
			fmt.Println("error while unmarshalling newsletter", err.Error())
			api.RespondWithJSON(rw, http.StatusBadRequest, "")
			return
		}
		api.RespondWithJSON(rw, http.StatusOK, "")

		newsletter.createContent(ctx)

		//fmt.Println("request", newsletter, ctx)
	})
}

func (s NewsLetter) createContent(ctx context.Context) {
	senderEmail := extractEmailBeforeLastDate(s.PlainText)
	if len(senderEmail) == 0 {
		fmt.Println("sender email not found")
		return
	}
	var id int64

	htmlText, err := s.removeElementsFromHTML(ctx)
	if err != nil {
		fmt.Println("error while parsing html", err.Error())
		return
	}
	plainText := s.parsePlainText()
	readTime := getReadTime(plainText)

	subject := strings.Replace(s.Header.Subject, "Fwd: ", "", 1)

	err = app.GetDB().GetContext(ctx, &id, setContent, htmlText, subject, senderEmail, plainText, readTime)
	if err == nil {
		go generateSummary(id)
	}

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"content_title_source_email\"" {
			fmt.Println("duplicate newsletter", id, err.Error(), "subject-", s.Header.Subject, "-sender-", senderEmail)
			_err := app.GetDB().GetContext(ctx, &id, getIDBySubjectAndSender, subject, senderEmail)
			if _err != nil {
				fmt.Println("error getting content by subject and sender, error: ", _err.Error())
				err = _err
				return
			}
		}
		fmt.Println("error creating content, error: ", err.Error())
		return
	}

	userID, err := s.getUserID(ctx)
	if err != nil {
		fmt.Println("error getting user id from newsletter", err.Error())
		return
	}

	cardID, err := card(userID, id)
	if err != nil {
		fmt.Println("error generating card", err.Error())
		return
	}

	fmt.Println("cardID", cardID)
}

func card(userID int64, contentID int64) (id int64, err error) {
	err = app.GetDB().GetContext(context.Background(), &id, createCard, userID, contentID, "active")
	return
}

func (s NewsLetter) getUserID(ctx context.Context) (id int64, err error) {
	err = app.GetDB().GetContext(ctx, &id, getUserIDFromEmail, s.Header.UserEmail)
	return
}

func getReadTime(text string) int {
	words := strings.Fields(text)
	readTime := int(len(words)) / 180
	return readTime
}

func generateSummary(id int64) (err error) {
	req := SummaryRequest{ID: id}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		fmt.Println("cant marshal", err.Error())
		return
	}

	client := httpclient.NewClient(httpclient.WithHTTPTimeout(30 * time.Second))

	resp, err := api.Post(context.Background(), "http://localhost:9223/summarize", reqJSON, nil, client)
	if err != nil {
		fmt.Println("error calling api", err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = errors.New("not 200 from ai server")
		return
	}
	return
}

func extractEmailBeforeLastDate(text string) string {
	// Split the text by "Date:"
	parts := strings.Split(text, "Date:")
	if len(parts) < 2 {
		fmt.Println("error wrror")
	}

	// Find the last occurrence of "Date:"
	lastDateIndex := strings.LastIndex(text, "Date:")

	// Extract the email address from the part before the last "Date:"
	beforeLastDate := text[:lastDateIndex]

	fmt.Println(beforeLastDate)
	// Define a regular expression pattern to match an email address
	pattern := `<[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}>`

	// Compile the regex pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Println("error regex")
	}

	// Find the first match in the part before the last "Date:"
	match := regex.FindAllString(beforeLastDate, -1)

	return match[len(match)-1][1 : len(match[len(match)-1])-1]
}
