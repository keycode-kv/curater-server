package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

type loginResponse struct {
	Password      string `json:"password"`
	RedirectEmail string `json:"redirect_email"`
}

type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userProfile struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	ArticleCount int64  `json:"article_count"`
}

type collectionInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type collectionList struct {
	Collections []collectionInfo `json:"collections"`
}

type userConfigResponse struct {
	RedirectEmail string `json:"redirect_email"`
}

func (s loginRequest) login(ctx context.Context) (resposne authResponse, err error) {
	inputPword := Hash(s.Password)

	userInfo, err := getLoginInfoByEmail(ctx, s.Email)
	if err != nil {
		fmt.Printf("error getting user: %s, error: %s\n", ctx.Value("user"), err.Error())
		return
	}

	if inputPword == userInfo.Password {
		tok := base64.StdEncoding.EncodeToString([]byte(s.Email))
		resposne = authResponse{
			Token: tok,
		}
	} else {
		err = errors.New("login failed")
	}
	return
}

func (s signupRequest) signup(ctx context.Context) (response authResponse, err error) {
	_, err = getLoginInfoByEmail(ctx, s.Email)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("error checking existing user: %s\n", err.Error())
		return
	}
	if err == nil {
		err = errors.New("user account already exists")
		return
	}

	_, err = createUser(ctx, s)
	if err != nil {
		fmt.Printf("error creating user: %s\n", err.Error())
		return
	}

	token := base64.StdEncoding.EncodeToString([]byte(s.Email))
	response = authResponse{
		Token: token,
	}
	return
}

func getProfile(ctx context.Context) (user userProfile, err error) {
	userInfo, err := getUserByEmail(ctx)
	if err != nil {
		fmt.Printf("error getting user: %s, error: %s\n", ctx.Value("user"), err.Error())
		return
	}

	articleCount, err := getArticleCountByUserID(ctx)
	if err != nil {
		fmt.Printf("error getting article count for user %s, error: %s\n", ctx.Value("user"), err.Error())
		return
	}

	user = userProfile{
		ID:           userInfo.ID,
		Name:         userInfo.Name,
		Email:        userInfo.Email,
		ArticleCount: articleCount,
	}

	return
}

func getCollections(ctx context.Context) (collections collectionList, err error) {
	userInfo, err := getUserByEmail(ctx)
	if err != nil {
		fmt.Printf("error getting user: %s, error: %s\n", ctx.Value("user"), err.Error())
		return
	}

	collectionList, err := getCollectionsByUserID(ctx, userInfo.ID)
	if err != nil {
		fmt.Printf("error getting article count for user %s, error: %s\n", ctx.Value("user"), err.Error())
		return
	}

	for _, item := range collectionList {
		collections.Collections = append(collections.Collections, collectionInfo{
			ID:   item.ID,
			Name: item.Name,
		})
	}

	return
}

func getuserConfig(ctx context.Context) (response userConfigResponse, err error) {
	redirectEmail, err := getUserConfigByEmail(ctx)
	if err != nil {
		fmt.Printf("error getting rediret email for user: %s, error: %s\n", ctx.Value("user"), err.Error())
		return
	}
	if redirectEmail != nil {
		response.RedirectEmail = *redirectEmail
	}
	return
}

func Hash(input string) string {
	hasher := sha256.New()

	hasher.Write([]byte(input))

	hashBytes := hasher.Sum(nil)

	hashString := hex.EncodeToString(hashBytes)

	return hashString
}
