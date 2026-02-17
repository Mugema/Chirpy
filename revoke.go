package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/Mugema/Chirpy/internal/auth"
	"github.com/Mugema/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerRevoke(writer http.ResponseWriter, req *http.Request) {
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		fmt.Printf("No token provided in header. Error: %v", err)
		writer.WriteHeader(401)
		return
	}

	err = cfg.db.RevokeToken(req.Context(),
		database.RevokeTokenParams{
			RevokedAt: sql.NullTime{Time: time.Now().Local(), Valid: true},
			UpdatedAt: time.Now().Local(),
			Token:     bearerToken})
	if err != nil {
		fmt.Printf("Error revoking token access %v", err)
		writer.WriteHeader(502)
		return
	}

	writer.WriteHeader(204)
	return

}
