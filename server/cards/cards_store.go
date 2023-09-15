package cards

import (
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
		    cards c inner join content ct on c.content_id = ct.id
			inner join users u on c.user_id = u.id
		WHERE
			u.email = $1
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

	getTagCl

	getTags = `SELECT t.tag
        FROM content_tags ct
        JOIN tags t ON ct.tag_id = t.id
        WHERE ct.content_id = $1`
)

type Card struct {
	ID            int      `db:"id" json:"id,omitempty"`
	ContentID     int      `db:"content_id" json:"-"`
	Title         string   `db:"title" json:"title,omitempty"`
	Content       string   `db:"content" json:"-"`
	Rating        float64  `json:"rating"`
	RatingCount   int      `json:"rating_count"`
	CommentsCount int      `json:"comments_count"`
	ViewCount     int      `db:"view_count" json:"view_count,omitempty"`
	SourceEmail   string   `db:"source_email" json:"source_email,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Duration      int      `json:"duration,omitempty"`
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
