package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"local/mda/internal/auth"
	"local/mda/internal/database"

	"github.com/google/uuid"
)

type User struct {
	Id uuid.UUID 		`json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email string 		`json:"email"`
	Token string		`json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type createUserReq struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := createUserReq{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	email := params.Email
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "There was a problem with your password", err)
	}

	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Email: email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Error creating user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		Id: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})
}