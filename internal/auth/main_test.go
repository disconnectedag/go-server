package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswordHashing(t *testing.T) {
	pw := "password"
	hash, _ := HashPassword(pw)
	pw_match, _ := CheckPasswordHash(pw, hash)
	if !pw_match {
		t.Fail()
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	parsedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	if parsedID != userID {
		t.Errorf("expected %v, got %v", userID, parsedID)
	}
}

func TestValidateJWT_Expired(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, -time.Hour) // already expired
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	_, err = ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		headers     http.Header
		expected    string
		expectError bool
	}{
		{
			name:        "valid bearer token",
			headers:     http.Header{"Authorization": []string{"Bearer my-jwt-token"}},
			expected:    "my-jwt-token",
			expectError: false,
		},
		{
			name:        "missing authorization header",
			headers:     http.Header{},
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty authorization header",
			headers:     http.Header{"Authorization": []string{""}},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)
			if tt.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, token)
			}
		})
	}
}