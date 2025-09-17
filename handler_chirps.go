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

type Chirp struct {
	Id uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`	
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
	UserId uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type createChirpRequest struct {
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
	respondWithJSON(w, http.StatusCreated, Chirp {
		Id: chirpEntity.ID,
		CreatedAt: chirpEntity.CreatedAt,
		UpdatedAt: chirpEntity.UpdatedAt,
		Body: chirpEntity.Body,
		UserId: chirpEntity.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(context.Background())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error when getting chirps: %w", err)
		return
	}

	chirps_converted := []Chirp{} // initialize an empty slice

	for _, chirp := range chirps {
		chirps_converted = append(chirps_converted, Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps_converted)
}