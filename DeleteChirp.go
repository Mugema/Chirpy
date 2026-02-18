package main

import (
	"fmt"
	"net/http"

	"github.com/Mugema/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerDeleteChirp(writer http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	userid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	chirpId, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		return
	}

	chirp, err := cfg.db.GetChirpById(req.Context(), chirpId)
	if err != nil {
		writer.WriteHeader(404)
		return
	}

	if chirp.UserID != userid {
		writer.WriteHeader(403)
		return
	}

	err = cfg.db.DeleteChirp(req.Context(), chirpId)
	if err != nil {
		fmt.Printf("Failed to delete chirp Error:%v \n", err)
		return
	}

	writer.WriteHeader(204)
	return

}
