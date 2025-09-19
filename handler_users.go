package main

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"local/mda/internal/auth"
	"local/mda/internal/database"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

type updateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	// 1) Require access token
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil || bearer == "" {
		respondWithError(w, http.StatusUnauthorized, "missing or invalid authorization header", nil)
		return
	}

	userID, err := auth.ValidateJWT(bearer, cfg.authSecret) // returns uuid.UUID if valid
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid or expired token", nil)
		return
	}

	// 2) Parse body (both fields required)
	var body updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't decode request body", err)
		return
	}
	if body.Email == "" || body.Password == "" {
		respondWithError(w, http.StatusBadRequest, "email and password are required", nil)
		return
	}

	// Minimal email check
	var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !emailRe.MatchString(body.Email) {
		respondWithError(w, http.StatusBadRequest, "invalid email format", nil)
		return
	}
	if len(body.Password) < 8 {
		respondWithError(w, http.StatusBadRequest, "password must be at least 8 characters", nil)
		return
	}

	// 3) Hash new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't hash password", err)
		return
	}

	// 4) Update DB (only the authenticated user)
	u, err := cfg.db.UpdateUserEmailAndPassword(
		context.Background(),
		database.UpdateUserEmailAndPasswordParams{
			ID:             userID,
			Email:          body.Email,
			HashedPassword: string(hashed),
		},
	)
	if err != nil {
		if isUniqueViolation(err) {
			respondWithError(w, http.StatusConflict, "email already in use", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "couldn't update user", err)
		return
	}

	// 5) Respond (omit password)
	respondWithJSON(w, http.StatusOK, userResponse{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	})
}

// Driver-specific helpers (implement for your driver)
func isUniqueViolation(err error) bool {
	// e.g., for lib/pq: pqErr.Code == "23505"
	return false
}
