package auth

import (
	"context"
	"curater/app"
	"time"
)

const (
	getLoginInfoByEmailQuery = `select password from users where email = $1`

	getUserConfigByEmailQuery = `select redirect_email from users where email = $1`

	createUserQuery = `INSERT INTO public.users
		(email, "name", "password", created_at, updated_at)
		VALUES($1, $2, $3, now(), now()) RETURNING id, email, name;`

	getUserInfoByEmail = `SELECT id, email, "name", redirect_email
		FROM users u WHERE email = $1;`

	getArticleCountByUserIDQuery = `SELECT count(c.id) as article_count
		FROM cards c INNER JOIN users u ON u.id = c.user_id WHERE u.email = $1`

	getCollectionsByUserIDQuery = `SELECT id, "name", user_id
		FROM collection where user_id = $1;`
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

type Collection struct {
	ID     int64  `json:"id" db:"id"`
	UserID int64  `json:"user_id" db:"user_id"`
	Name   string `json:"name" db:"name"`
}

func getLoginInfoByEmail(ctx context.Context, email string) (userInfo user, err error) {
	err = app.GetDB().GetContext(ctx, &userInfo, getLoginInfoByEmailQuery, email)
	if err != nil {
		return
	}
	return
}

func createUser(ctx context.Context, request signupRequest) (userInfo user, err error) {
	inputPword := Hash(request.Password)

	err = app.GetDB().GetContext(ctx, &userInfo, createUserQuery, request.Email, request.Name, inputPword)
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

func getArticleCountByUserID(ctx context.Context) (articleCount int64, err error) {
	err = app.GetDB().GetContext(ctx, &articleCount, getArticleCountByUserIDQuery, ctx.Value("user"))
	if err != nil {
		return
	}

	return
}

func getCollectionsByUserID(ctx context.Context, userID int64) (collections []Collection, err error) {
	err = app.GetDB().SelectContext(ctx, &collections, getCollectionsByUserIDQuery, userID)
	if err != nil {
		return
	}
	return
}

func getUserConfigByEmail(ctx context.Context) (redirectEmail *string, err error) {
	err = app.GetDB().GetContext(ctx, &redirectEmail, getUserConfigByEmailQuery, ctx.Value("user"))
	if err != nil {
		return
	}
	return
}
