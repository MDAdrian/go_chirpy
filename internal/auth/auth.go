package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)


func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil	
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// Create claims
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}

	// Create a new token with HS256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			// Enforce HS256 explicitly
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(tokenSecret), nil
		},
		// Extra guard so non-HS256 tokens fail fast
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

	// Optional: enforce issuer if you want stricter validation
	if claims.Issuer != "chirpy" {
		return uuid.UUID{}, fmt.Errorf("unexpected issuer: %s", claims.Issuer)
	}

	uid, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid subject uuid: %w", err)
	}
	return uid, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	const prefix = "Bearer "

	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header not found")
	}

	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("authorization header is not a Bearer token")
	}

	// Strip the "Bearer " prefix and trim whitespace
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	if token == "" {
		return "", errors.New("token string is empty")
	}

	return token, nil
}

func MakeRefreshToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits

	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("could not generate random bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

var ErrNoAuthHeaderIncluded = errors.New("no authorization header included")

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}

	// Expect "ApiKey THE_KEY_HERE"
	const prefix = "ApiKey "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("malformed authorization header")
	}

	return strings.TrimSpace(strings.TrimPrefix(authHeader, prefix)), nil
}
