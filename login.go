package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Mugema/Chirpy/internal/auth"
	"github.com/Mugema/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(writer http.ResponseWriter, req *http.Request) {
	type authDetails struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	userAuth := authDetails{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&userAuth)
	if err != nil {
		fmt.Println("Error Decoding")
		return
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), userAuth.Email)

	if err != nil {
		writer.WriteHeader(401)
		return
	}

	isPasswordValid, err := auth.CheckPassword(userAuth.Password, user.Password)
	if err != nil {
		fmt.Printf("Error while confirming password, The error:%v", err)
		writer.WriteHeader(401)
		return
	}

	mappedUser := jsonUserMapper(user)
	if isPasswordValid {
		ExpiresIn := 3600
		token, err := auth.MakeJWT(mappedUser.ID, cfg.secret, time.Duration(ExpiresIn)*time.Second)
		if err != nil {
			fmt.Printf("Error while making the token, %v", err)
		}
		refreshToken, err := auth.MakeRefreshToken()

		_, err = cfg.db.CreateToken(
			req.Context(),
			database.CreateTokenParams{
				Token:     refreshToken,
				CreatedAt: time.Now().Local(),
				UpdatedAt: time.Now().Local(),
				RevokedAt: sql.NullTime{},
				ExpiresAt: (time.Now().Local()).Add(time.Duration(1440) * time.Hour),
				UserID:    mappedUser.ID,
			})
		if err != nil {
			fmt.Printf("Error generating the refresh token , %v", err)
			return
		}

		data, err := json.Marshal(
			struct {
				User
				Token        string `json:"token"`
				RefreshToken string `json:"refresh_token"`
			}{
				mappedUser,
				token,
				refreshToken})
		if err != nil {
			fmt.Println("Error marshalling")
			return
		}

		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(http.StatusOK)

		writer.Write(data)
		return
	}

	writer.WriteHeader(401)
	return

}
