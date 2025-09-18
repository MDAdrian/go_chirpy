// auth/auth_test.go
package tests

import (
	"net/http"
	"testing"
	"time"

	"local/mda/internal/auth"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT_Success(t *testing.T) {
	secret := "topsecret"
	userID := uuid.New()
	token, err := auth.MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	gotID, err := auth.ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT error: %v", err)
	}
	if gotID != userID {
		t.Fatalf("expected %s, got %s", userID, gotID)
	}
}

func TestValidateJWT_Expired(t *testing.T) {
	secret := "topsecret"
	userID := uuid.New()

	// Expire immediately by using a negative duration
	token, err := auth.MakeJWT(userID, secret, -1*time.Second)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	if _, err := auth.ValidateJWT(token, secret); err == nil {
		t.Fatalf("expected error for expired token, got nil")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()
	token, err := auth.MakeJWT(userID, "right-secret", time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	if _, err := auth.ValidateJWT(token, "wrong-secret"); err == nil {
		t.Fatalf("expected error for wrong secret, got nil")
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name: "valid bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer abc123"},
			},
			wantToken: "abc123",
			wantErr:   false,
		},
		{
			name:      "no authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "invalid prefix",
			headers: http.Header{
				"Authorization": []string{"Basic abc123"},
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "empty token",
			headers: http.Header{
				"Authorization": []string{"Bearer "},
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := auth.GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}

