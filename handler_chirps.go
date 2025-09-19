package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"local/mda/internal/auth"
	"local/mda/internal/database"
	"net/http"
	"sort"
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
	}

	decoder := json.NewDecoder(r.Body)
	params := createChirpRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to access", nil)
		return
	}

	userId, err := auth.ValidateJWT(bearer, cfg.authSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to access", nil)
		return
	}

	chirp := cleanProfaneWords(params.Body)

	chirpEntity, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{
		Body: chirp,
		UserID: userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating chirp: %w", err)
		return
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

func (cfg *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpId")
	chirpUuid, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error when decoding path value id ", err)
		return
	}

	chirp, err := cfg.db.GetChirpById(context.Background(), chirpUuid)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "could not find chirp with that id", nil)
		return
	}
	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error when getting chirp by id", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}


func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	// 1) Require and validate access token
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil || bearer == "" {
		respondWithError(w, http.StatusUnauthorized, "missing or invalid authorization header", nil)
		return
	}
	userID, err := auth.ValidateJWT(bearer, cfg.authSecret) // returns uuid.UUID for the subject
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid or expired token", nil)
		return
	}

	// 2) Parse chirp ID from path
	idStr := r.PathValue("chirpId") // make sure your route uses {chirpID}; keep the casing consistent
	chirpUUID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp id", err)
		return
	}

	// 3) Load chirp (404 if not found)
	chirp, err := cfg.db.GetChirpById(context.Background(), chirpUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "chirp not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "couldn't fetch chirp", err)
		return
	}

	// 4) Ownership check (403 if not author)
	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "you are not allowed to delete this chirp", nil)
		return
	}

	// 5) Delete (204 on success)
	if err := cfg.db.DeleteChirp(context.Background(), chirpUUID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't delete chirp", err)
		return
	}
	w.WriteHeader(http.StatusNoContent) // 204
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authorParam := r.URL.Query().Get("author_id")
	var (
		rows []database.Chirp
		err  error
	)

	if authorParam == "" {
		// No filter → all chirps, ASC by created_at
		rows, err = cfg.db.GetChirps(ctx)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "couldn't fetch chirps", err)
			return
		}
	} else {
		// Filter by author_id (UUID expected)
		authorID, parseErr := uuid.Parse(authorParam)
		if parseErr != nil {
			respondWithError(w, http.StatusBadRequest, "invalid author_id format (expected UUID)", parseErr)
			return
		}

		rows, err = cfg.db.GetChirpsByAuthor(ctx, authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "couldn't fetch chirps for author", err)
			return
		}
	}

	// Map DB → API shape (omit sensitive fields)
	out := make([]Chirp, 0, len(rows))
	for _, c := range rows {
		out = append(out, Chirp{
			Id:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserId:    c.UserID,
		})
	}

	// 3) Check query param "sort" (default asc)
	sortParam := r.URL.Query().Get("sort")
	if sortParam != "desc" {
		sortParam = "asc"
	}

	sort.Slice(out, func(i, j int) bool {
		if sortParam == "asc" {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[j].CreatedAt.Before(out[i].CreatedAt)
	})

	// 4) Respond
	respondWithJSON(w, http.StatusOK, out)
}
