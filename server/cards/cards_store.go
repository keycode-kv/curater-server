package cards

import (
	"context"
	"curater/app"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"strings"
)

const (
	getCards = `        
		SELECT 
		    c.id,
		    ct.id as content_id,
		    ct.title,
		    ct.content,
			ct.view_count,
			ct.source_email
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

	getCardByID = `
			WITH CardInfo AS (
				SELECT
					c.id AS card_id,
					c.content_id,
					u.email AS user_email
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
				COALESCE(r.average_rating, 0) AS rating,
				COALESCE(cm.total_comment_count, 0) AS comment_count
			FROM
				CardInfo ci
			LEFT JOIN
				content cnt ON ci.content_id = cnt.id
			LEFT JOIN
				RatingInfo r ON ci.card_id = r.content_id
			LEFT JOIN
				CommentInfo cm ON cnt.id = cm.content_id
		`

	getCommentsByID = `
			SELECT
				c.id,
				c.comment AS content,
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

	getSearchClause = ` AND (
			ct.title ILIKE '%s' OR
			ct.content ILIKE '%s' OR
			ct.summary ILIKE '%s' 
			)`

	getTags = `SELECT t.tag
        FROM content_tags ct
        JOIN tags t ON ct.tag_id = t.id
        WHERE ct.content_id = $1`

	getTagListQuery = `select tag from tags;`

	updateCardByIDQuery = `UPDATE cards
		SET updated_at = now()`

	updateCardByIDWhereClause = ` WHERE id = $1 RETURNING id, status;`

	updateCardStats      = `, status = $%d`
	updateCardCollection = `, collection_id = $%d`
	updateCardIsViewed   = `, is_viewed = $%d`
)

type Card struct {
	ID            int      `db:"id" json:"id,omitempty"`
	ContentID     int      `db:"content_id" json:"-"`
	Title         string   `db:"title" json:"title,omitempty"`
	Content       string   `db:"content" json:"-"`
	Status        string   `db:"status" json:"-"`
	Rating        float64  `json:"rating"`
	RatingCount   int      `json:"rating_count"`
	CommentsCount int      `json:"comments_count"`
	ViewCount     int      `db:"view_count" json:"view_count,omitempty"`
	SourceEmail   string   `db:"source_email" json:"source_email,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Duration      int      `json:"duration,omitempty"`
}

type ContentData struct {
	ContentID    string  `db:"id" json:"content_id"`
	Content      string  `db:"content" json:"content"`
	Title        string  `db:"title" json:"title"`
	Rating       float64 `db:"rating" json:"rating"`
	CommentCount int64   `db:"comment_count" json:"comment_count"`
}

type Comment struct {
	ID          string `db:"id" json:"id,omitempty"`
	Content     string `db:"content" json:"content,omitempty"`
	User        string `db:"user" json:"user,omitempty"`
	CommentedAt int64  `db:"commented_at" json:"commented_at,omitempty"`
}

func GetCardsForUser(userID string, filters Filter) ([]Card, error) {
	var cards []Card

	query := getCards

	typeClause := fmt.Sprintf(getTypeClause, "active")

	if len(filters.Type) != 0 {
		typeClause = fmt.Sprintf(getTypeClause, filters.Type)
	}
	query = query + typeClause

	if len(filters.Search) != 0 {
		filters.Search = "%" + filters.Search + "%"
		query = query + fmt.Sprintf(getSearchClause, filters.Search, filters.Search, filters.Search)
	}

	err := app.GetDB().Select(&cards, query, userID)
	if err != nil {
		return nil, err
	}

	for i, card := range cards {
		tags, err := GetTagsForCard(card.ID)
		if err != nil {
			return nil, err
		}

		cards[i].Rating, cards[i].RatingCount, err = GetRatingForContent(card.ContentID)
		if err != nil {
			return nil, err
		}

		cards[i].CommentsCount, err = GetCommentsForContent(card.ContentID)
		if err != nil {

			return nil, err
		}

		cards[i].Tags = tags
		cards[i].Duration = EstimateReadTime(cards[i].Content, 200)
		if len(filters.Tags) > 0 {
			cardTagMap := make(map[string]bool)
			for _, tag := range tags {
				cardTagMap[tag] = true
			}
			var isNotMatch bool
			for _, tag := range filters.Tags {
				if !cardTagMap[tag] {
					isNotMatch = true
				}
			}
			if isNotMatch {
				cards[i] = Card{}
			}
		}
	}

	return cards, nil
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

func GetCardByIDForUser(userID string, cardID string) (card ContentData, err error) {
	err = app.GetDB().Get(&card, getCardByID, userID, cardID)
	return
}

func GetCommentsByContentID(userID string, contentID string) (comments []Comment, err error) {
	err = app.GetDB().Select(&comments, getCommentsByID, userID, contentID)
	return
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

	// fmt.Println("query: ", query)
	// fmt.Println("argsList: ", argsList)
	err = app.GetDB().GetContext(ctx, &cardInfo, query, argsList...)
	if err != nil {
		return
	}
	return
}
