package main

import (
	"context"
	"encoding/json"
	"fmt"
	"local/mda/internal/database"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type createChirpRequest struct {
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	type createChirpResponse struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`	
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := createChirpRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	chirp := cleanProfaneWords(params.Body)

	chirpEntity, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{
		Body: chirp,
		UserID: params.UserId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating chirp: %w", err)
	}

	fmt.Printf("Created chirp for user: %s, message: '%s'\n", chirpEntity.UserID, chirpEntity.Body)
	respondWithJSON(w, http.StatusCreated, createChirpResponse {
		Id: chirpEntity.ID,
		CreatedAt: chirpEntity.CreatedAt,
		UpdatedAt: chirpEntity.UpdatedAt,
		Body: chirpEntity.Body,
		UserId: chirpEntity.UserID,
	})
}