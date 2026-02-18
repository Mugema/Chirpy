package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Mugema/Chirpy/internal/auth"
	"github.com/Mugema/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerUpdateUser(writer http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	p := parameters{}
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&p)
	if err != nil {
		fmt.Printf("Error decoding body \n")
		return
	}

	hashString, err := auth.HashPassword(p.Password)
	if err != nil {
		return
	}

	user, err := cfg.db.UpdateUser(req.Context(),
		database.UpdateUserParams{
			Email:     p.Email,
			Password:  hashString,
			UpdatedAt: time.Now().Local(),
			ID:        userId})

	if err != nil {
		fmt.Printf("Error updating the user. Error:%v\n", err)
		return
	}

	data, err := json.Marshal(jsonUserMapper(user))

	writer.Write(data)
	return
}
