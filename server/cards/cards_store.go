package cards

import (
	"context"
	"curater/app"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"
)

const (
	getCards = `        
		SELECT 
		    c.id,
		    c.collection_id,
		    ct.id as content_id,
		    ct.title,
		    ct.content,
			ct.source_email,
			ct.summary
		FROM
		    cards c 
		INNER JOIN 
		    content ct on c.content_id = ct.id
		INNER JOIN 
		    users u on c.user_id = u.id
		WHERE
			u.email = $1 AND
			ct.summary IS NOT NULL
        `

	getContentViewedCount = `SELECT count(id) FROM cards WHERE is_viewed = true AND content_id = $1;`

	getCardByID = `
			WITH CardInfo AS (
				SELECT
					c.id AS card_id,
					c.content_id,
					u.email AS user_email,
					c.is_viewed
				FROM
					cards c
				JOIN
					users u ON c.user_id = u.id
				WHERE
					u.email = $1
					AND c.id = $2
			),
			RatingInfo AS (
				SELECT
					r.content_id,
					COALESCE(AVG(r.rating), 0) AS average_rating
				FROM
					rating r
				GROUP BY
					r.content_id
			),
			CommentInfo AS (
				SELECT
					c.content_id,
					COUNT(c.id) AS total_comment_count
				FROM
					comments c
				GROUP BY
					c.content_id
			)
			SELECT
			    cnt.id,
				cnt.content,
				cnt.title,
				ci.is_viewed,
				COALESCE(r.average_rating, 0) AS rating,
				COALESCE(cm.total_comment_count, 0) AS comment_count
			FROM
				CardInfo ci
			LEFT JOIN
				content cnt ON ci.content_id = cnt.id
			LEFT JOIN
				RatingInfo r ON ci.content_id = r.content_id
			LEFT JOIN
				CommentInfo cm ON cnt.id = cm.content_id
		`

	getCommentsByID = `
			SELECT
				c.id,
				c.comment,
				u.name AS user,
				CAST(EXTRACT(EPOCH FROM c.created_at)::int AS int) AS commented_at
			FROM
				comments c
			JOIN
				users u ON c.user_id = u.id
			WHERE
			    u.email = $1 AND
				c.content_id = $2
			ORDER BY
				c.created_at ASC
		`

	getRatingByContent = `
		SELECT
		    AVG(rating) AS average_rating,
		    COUNT(*) AS user_count
		FROM
		    rating
		WHERE
		    content_id = $1
		GROUP BY
		    content_id;
	`

	getCommentsByContent = `		SELECT
		    COUNT(*) AS comment_count
		FROM
		    comments
		WHERE
		    content_id = $1
		GROUP BY
		    content_id;`

	getTypeClause = ` AND
			c.status = '%s'`

	getCollectionClause = ` AND
			c.collection_id = %s`

	getSearchClause = ` AND (
			ct.title ILIKE '%s' OR
			ct.summary ILIKE '%s' 
			)`

	getTags = `SELECT t.tag
        FROM content_tags ct
        JOIN tags t ON ct.tag_id = t.id
        WHERE ct.content_id = $1`

	getUserFromEmail = `
		SELECT id FROM users
		WHERE email = $1
`

	insertRating = ` 
		INSERT INTO rating (user_id, content_id, rating, created_at, updated_at) VALUES($1, $2, $3, now(), now())
		        ON CONFLICT (user_id, content_id)
        DO UPDATE SET rating = EXCLUDED.rating, updated_at = NOW()
`
	getTagListQuery = `select tag from tags;`

	updateCardByIDQuery = `UPDATE cards
		SET updated_at = now()`

	updateCardByIDWhereClause = ` WHERE id = $1 RETURNING id, status;`

	updateCardStats      = `, status = $%d`
	updateCardCollection = `, collection_id = $%d`
	updateCardIsViewed   = `, is_viewed = $%d`

	createCommentQuery = `INSERT INTO "comments"
		(user_id, content_id, "comment", created_at, updated_at)
		VALUES($1, $2, $3, now(), now()) RETURNING id, comment,
		CAST(EXTRACT(EPOCH FROM created_at)::int AS int) AS commented_at;`

	getUserInfoByEmail = `SELECT id, email, "name", redirect_email
		FROM users u WHERE email = $1;`
)

type user struct {
	ID            int64     `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	Name          string    `json:"name" db:"name"`
	Password      string    `json:"password" db:"password"`
	RedirectEmail *string   `json:"redirect_email" db:"redirect_email"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
type Card struct {
	ID            int      `db:"id" json:"id"`
	CollectionID  *int     `db:"collection_id" json:"collection_id"`
	ContentID     int      `db:"content_id" json:"content_id"`
	Title         string   `db:"title" json:"title"`
	Content       string   `db:"content" json:"-"`
	Status        string   `db:"status" json:"-"`
	Summary       string   `db:"summary" json:"summary"`
	Rating        float64  `json:"rating"`
	RatingCount   int      `json:"rating_count"`
	CommentsCount int      `json:"comments_count"`
	ViewCount     int      `db:"view_count" json:"view_count"`
	SourceEmail   string   `db:"source_email" json:"source_email"`
	Tags          []string `json:"tags"`
	Duration      int      `json:"duration"`
}

type ContentData struct {
	ContentID    string  `db:"id" json:"content_id"`
	Content      string  `db:"content" json:"content"`
	Title        string  `db:"title" json:"title"`
	Rating       float64 `db:"rating" json:"rating"`
	CommentCount int64   `db:"comment_count" json:"comment_count"`
	IsViewed     bool    `db:"is_viewed" json:"-"`
}

type Comment struct {
	ID          string `db:"id" json:"id,omitempty"`
	Comment     string `db:"comment" json:"comment,omitempty"`
	User        string `db:"user" json:"user,omitempty"`
	CommentedAt int64  `db:"commented_at" json:"commented_at,omitempty"`
}

func GetCardsForUser(userID string, filters Filter) (response Cards, err error) {
	var cards []Card
	response = Cards{
		Cards: []Card{},
	}

	query := getCards

	typeClause := fmt.Sprintf(getTypeClause, "active")

	if len(filters.Type) != 0 {
		typeClause = fmt.Sprintf(getTypeClause, filters.Type)
	}
	query = query + typeClause

	if len(filters.Collection) != 0 {
		query = query + fmt.Sprintf(getCollectionClause, filters.Collection)
	}
	cardTagMap := make(map[string]bool)
	if len(filters.Tags) > 0 {
		for _, tag := range filters.Tags {
			cardTagMap[tag] = true
		}
	}

	if len(filters.Search) != 0 {
		filters.Search = "%" + filters.Search + "%"
		query = query + fmt.Sprintf(getSearchClause, filters.Search, filters.Search)
	}

	err = app.GetDB().Select(&cards, query, userID)
	if err != nil {
		fmt.Println("error selecting cards, error: ", err.Error())
		return
	}
	if len(cards) > 0 {
		response.Cards = cards
	}
	filteredCards := []Card{}

	for i, card := range cards {
		var viewCount int
		err = app.GetDB().Get(&viewCount, getContentViewedCount, card.ContentID)
		if err != nil {
			fmt.Printf("error getting viewed count, error: %s, content_id: %d", err.Error(), card.ContentID)
			return
		}
		cards[i].ViewCount = viewCount

		tags, err := GetTagsForCard(card.ID)
		if err != nil {
			return response, err
		}

		cards[i].Rating, cards[i].RatingCount, err = GetRatingForContent(card.ContentID)
		if err != nil {
			return response, err
		}

		cards[i].CommentsCount, err = GetCommentsForContent(card.ContentID)
		if err != nil {
			return response, err
		}
		cards[i].Tags = tags
		cards[i].Duration = EstimateReadTime(cards[i].Content, 200)
		if len(filters.Tags) > 0 {
			var isMatch bool
			for _, tag := range tags {
				if cardTagMap[tag] {
					isMatch = true
					break
				}
			}
			if isMatch {
				filteredCards = append(filteredCards, cards[i])
			}
		}
	}

	if len(filters.Tags) > 0 {
		response.Cards = filteredCards
	}

	return response, nil
}

func GetTagsForCard(cardID int) ([]string, error) {
	var tags []string

	err := app.GetDB().Select(&tags, getTags, cardID)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func GetRatingForContent(contentID int) (float64, int, error) {
	var averageRating float64
	var userCount int

	err := app.GetDB().QueryRow(getRatingByContent, contentID).Scan(&averageRating, &userCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		return 0, 0, err
	}

	return averageRating, userCount, nil
}

func GetCommentsForContent(contentID int) (int, error) {
	var commentCount int

	err := app.GetDB().QueryRow(getCommentsByContent, contentID).Scan(&commentCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		return 0, err
	}

	return commentCount, nil
}

func EstimateReadTime(htmlContent string, wordsPerMinute int) int {
	// Remove HTML tags and decode HTML entities
	textContent := html.UnescapeString(stripHTMLTags(htmlContent))

	// Split the content into words
	words := strings.Fields(textContent)

	// Calculate the estimated read time based on the number of words and words per minute
	wordCount := len(words)
	readTime := wordCount / wordsPerMinute

	// Ensure a minimum read time of 1 minute
	if readTime < 1 {
		readTime = 1
	}

	return readTime
}

func GetCardByIDForUser(ctx context.Context, userEmail string, cardID int64) (card ContentData, err error) {
	err = app.GetDB().Get(&card, getCardByID, userEmail, cardID)
	if err != nil {
		fmt.Println("error getting card by id, error: ", err.Error())
		return
	}
	fmt.Println("CARD: ", card.IsViewed)
	if !card.IsViewed {
		_, err = updateCardViewStatus(ctx, cardID)
		if err != nil {
			fmt.Println("error updating card viewed status, error: ", err.Error())
			return
		}
	}
	return
}

func getCommentsByContentID(userID string, contentID string) (comments []Comment, err error) {
	err = app.GetDB().Select(&comments, getCommentsByID, userID, contentID)
	return
}

func InsertRating(userEmail string, request PostRatingRequest) (err error) {
	var userID int
	err = app.GetDB().QueryRow(getUserFromEmail, userEmail).Scan(&userID)
	if err != nil {
		fmt.Println("error getting user")
		return
	}
	_, err = app.GetDB().Exec(insertRating, userID, request.ContentID, request.Rating)
	if err != nil {
		fmt.Println("error inserting rating")
		return
	}
	return nil
}

// stripHTMLTags removes HTML tags from a string.
func stripHTMLTags(s string) string {
	var result string
	var withinTag bool

	for _, r := range s {
		if r == '<' {
			withinTag = true
			continue
		} else if r == '>' && withinTag {
			withinTag = false
			continue
		}

		if !withinTag {
			result += string(r)
		}
	}
	result = strings.Join(strings.Fields(result), " ")
	return result
}

func getTagList(ctx context.Context) (tags []string, err error) {
	err = app.GetDB().SelectContext(ctx, &tags, getTagListQuery)
	if err != nil {
		return
	}
	return
}

func updateCardInfo(ctx context.Context, request updateCardRequest) (cardInfo Card, err error) {
	var argsList []interface{}
	query := updateCardByIDQuery
	argsList = append(argsList, request.ID)

	if len(request.Status) > 0 {
		argsList = append(argsList, request.Status)
		query += fmt.Sprintf(updateCardStats, len(argsList))
	}
	if request.CollectionID != 0 {
		argsList = append(argsList, request.CollectionID)
		query += fmt.Sprintf(updateCardCollection, len(argsList))
	}
	if request.IsViewed {
		argsList = append(argsList, request.IsViewed)
		query += fmt.Sprintf(updateCardIsViewed, len(argsList))
	}
	query += updateCardByIDWhereClause

	err = app.GetDB().GetContext(ctx, &cardInfo, query, argsList...)
	if err != nil {
		return
	}
	return
}

func createComment(ctx context.Context, request commentRequest) (resp Comment, err error) {
	err = app.GetDB().GetContext(ctx, &resp, createCommentQuery, request.UserID, request.ContentID, request.Comment)
	if err != nil {
		return
	}
	return
}

func getUserByEmail(ctx context.Context) (userInfo user, err error) {
	err = app.GetDB().GetContext(ctx, &userInfo, getUserInfoByEmail, ctx.Value("user"))
	if err != nil {
		return
	}
	return
}

func updateCardViewStatus(ctx context.Context, cardID int64) (cardInfo Card, err error) {
	var argsList []interface{}
	query := updateCardByIDQuery
	argsList = append(argsList, cardID, true)
	query += fmt.Sprintf(updateCardIsViewed, len(argsList))
	query += updateCardByIDWhereClause

	err = app.GetDB().GetContext(ctx, &cardInfo, query, argsList...)
	if err != nil {
		return
	}
	return
}
