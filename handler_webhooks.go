package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"local/mda/internal/auth"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type requestWebhook struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	ctx := context.Background()

	// 1. Validate API key
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polkaKey { 
		respondWithError(w, http.StatusUnauthorized, "invalid API key", nil)
		return
	}

	// 2) Parse body (both fields required)
	var body requestWebhook
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't decode request body", err)
		return
	}

	if body.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}	

	_, err = cfg.db.SetUserToChirpyRed(ctx, body.Data.UserId)
	switch {
	case err == sql.ErrNoRows:
		respondWithError(w, http.StatusNotFound, "user not found", nil)
		return
	case err != nil:
		respondWithError(w, http.StatusInternalServerError, "failed to update user", err)
		return
	default:
		w.WriteHeader(http.StatusNoContent) // success, empty body
		return
	}
}