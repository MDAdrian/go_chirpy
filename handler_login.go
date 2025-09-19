package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"local/mda/internal/auth"
	"local/mda/internal/database"
	"net/http"
	"time"
)


func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := loginRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
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
		time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create token", err)
		return
	}

	// refresh tokens
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not build refresh token", err)
	}

	expiresAt := time.Now().UTC().Add(60 * 24 * time.Hour)
	refreshTokenEntity, err := cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: user.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		Id: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Token: token,
		RefreshToken: refreshTokenEntity.Token,
	})
}

func (cfg *apiConfig) hanldlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type refreshToken struct {
		Token string `json:"token"`
	}

	ctx := context.Background()

	token, err := auth.GetBearerToken(r.Header)
	if err != nil || token == "" {
		respondWithError(w, http.StatusInternalServerError, "missing authorization header", err)
		return
	}

	rt, err := cfg.db.GetUserFromRefreshToken(ctx, token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting refresh token from db", err)
		return
	}

	if rt.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "refresh token revoked", nil)
		return
	}
	now := time.Now().UTC()
	if !rt.ExpiresAt.After(now) {
		respondWithError(w, http.StatusUnauthorized, "refresh token expired", nil)
		return
	}

	accesToken, err := auth.MakeJWT(rt.UserID, cfg.authSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error when creating new refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, refreshToken {
		Token: accesToken,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil || refreshToken == "" {
		respondWithError(w, http.StatusUnauthorized, "missing or invalid authorization header", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(ctx, refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error when revoking token", err)
	}

	w.WriteHeader(http.StatusNoContent)
}