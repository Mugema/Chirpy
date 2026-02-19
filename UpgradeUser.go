package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Mugema/Chirpy/internal/auth"
	"github.com/Mugema/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpgradeUser(writer http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}

	params := parameters{}

	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		writer.WriteHeader(401)
		return
	}

	if apiKey != cfg.polkaKey {
		writer.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error while decoding Error:%v\n", err)
		writer.WriteHeader(404)
		return
	}

	if params.Event != "user.upgraded" {
		writer.WriteHeader(204)
		return
	}

	id, err := uuid.Parse(params.Data.UserId)
	if err != nil {
		fmt.Printf("%v\n", err)
		writer.WriteHeader(404)
		return
	}
	err = cfg.db.UpgradeUser(req.Context(),
		database.UpgradeUserParams{
			IsChirpyRed: true,
			UpdatedAt:   time.Now().Local(),
			ID:          id})
	if err != nil {
		writer.WriteHeader(404)
		return
	}

	writer.WriteHeader(204)
	return

}
