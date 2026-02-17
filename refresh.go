package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Mugema/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(writer http.ResponseWriter, req *http.Request) {

	type parameters struct {
		Token string `json:"token"`
	}
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		fmt.Printf("No token provided in header. Error: %v", err)
		writer.WriteHeader(401)
		return
	}

	token, err := cfg.db.GetToken(req.Context(), bearerToken)
	fmt.Println(token)
	if err != nil {
		fmt.Printf("Handler Refresh.\n Error getting token %v\n", err)
		writer.WriteHeader(401)
		return
	}

	if token.RevokedAt.Valid {
		fmt.Printf("Handler Refresh.\n token revoked \n")
		writer.WriteHeader(401)
		return
	}

	if (time.Now().Local()).After(token.ExpiresAt) {
		fmt.Printf("Handler Refresh.\n Token expired %v\n", err)
		writer.WriteHeader(401)
		return
	}

	accessToken, err := auth.MakeJWT(token.UserID, cfg.secret, 3600*time.Second)
	if err != nil {
		fmt.Printf("Error creating JWT %v", err)
		writer.WriteHeader(401)
		return
	}

	data, err := json.Marshal(parameters{accessToken})
	if err != nil {
		fmt.Printf("Error marshaling the data %v", err)
		writer.WriteHeader(401)
		return
	}

	writer.WriteHeader(200)
	writer.Write(data)
	return
}
