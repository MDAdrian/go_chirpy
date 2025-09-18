package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"local/mda/internal/auth"
	"net/http"
	"time"
)


func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Email string `json:"email"`
		Password string `json:"password"`
		ExpiresInSeconds *int `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := loginRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// expiration logic
	const maxExpiry = time.Hour
	expiry := maxExpiry // default 1 hour
	if params.ExpiresInSeconds != nil {
		// clamp to maximum of 1 hour
		requested := time.Duration(*params.ExpiresInSeconds) * time.Second
		if requested < maxExpiry {
			expiry = requested
		}
	}

	user, err := cfg.db.GetUserByEmail(context.Background(), params.Email)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "There was an issue getting a user by email", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	// make JWT
	token, err := auth.MakeJWT(
		user.ID,
		cfg.authSecret,
		expiry,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create token: %v", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		Id: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Token: token,
	})
}