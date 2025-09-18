// auth/auth_test.go
package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"local/mda/internal/auth"
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
