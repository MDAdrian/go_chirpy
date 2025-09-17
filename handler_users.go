package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type createUserReq struct {
		Email string `json:"email"`
	}

	type createUseResp struct {
		Id uuid.UUID 		`json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string 		`json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := createUserReq{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	email := params.Email

	user, err := cfg.db.CreateUser(context.Background(), email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Error creating user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, createUseResp{
		Id: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})
}