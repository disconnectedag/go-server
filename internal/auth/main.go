package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")
	if apiKey == "" {
		return "", errors.New("Authorization not present")
    }
    if !strings.HasPrefix(apiKey, "ApiKey ") {
		return "", errors.New("authorization header must use  scheme")
    }
	apiKey = strings.TrimPrefix(apiKey, "ApiKey ")

	log.Printf("api key is: %s", apiKey)
	return apiKey, nil
}

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "chirpy-access", IssuedAt: jwt.NewNumericDate(time.Now().UTC()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()), Subject: userID.String()})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil // must return same type used to sign
	})

	if err != nil {
		return uuid.Nil, err
	}
	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authString := headers.Get("Authorization")
	if authString == "" {
        return "", errors.New("Bearer JWT token not present")
    }
    if !strings.HasPrefix(authString, "Bearer ") {
        return "", errors.New("authorization header must use Bearer scheme")
    }
	if authString == "" {
		return "", errors.New("Bearer JWT token not present")
	}
	token := strings.TrimPrefix(authString, "Bearer ")
	return token, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	randStr := hex.EncodeToString(key)
	return randStr
}