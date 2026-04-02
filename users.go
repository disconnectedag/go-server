package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/disconnectedag/go-server/internal/auth"
	"github.com/disconnectedag/go-server/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	IsChirpyRed bool    `json:"is_chirpy_red"`
}

func (apiCfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// type ValidResponse struct {
	// 	User
	// 	Token string `json:"token"`
	// 	RefreshToken string `json:"refresh_token"`
	// }
	// }
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No Refresh Token in Headers", err)
	}
	rTokenItem, err := apiCfg.db.GetUserFromRefreshToken(r.Context(), rToken)
	if rTokenItem.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has been revoked", nil)
		return
	}

	expiryTime := 3600

	token, err := auth.MakeJWT(rTokenItem.UserID, apiCfg.jwtSecret, time.Duration(expiryTime)*time.Second)

	respondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (apiCfg *apiConfig) revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No Refresh Token in Headers", err)
	}
	err = apiCfg.db.RevokeRefreshToken(r.Context(), rToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "revoking internal error", err)
	}
	w.WriteHeader(http.StatusNoContent)

}
func (apiCfg *apiConfig) updatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type ValidResponse struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Parsing error", err)
	}
	jwt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "missing or invalid token", err)
		return
	}

	userID, err := auth.ValidateJWT(jwt, apiCfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token", err)
		return
	}
	newPw, err := auth.HashPassword(params.Password)

	user, err := apiCfg.db.UpdateUser(r.Context(), database.UpdateUserParams{ID: userID, Email: params.Email, HashedPassword: newPw})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error uddating user", err)
		return
	}
	respondWithJSON(w, http.StatusOK, ValidResponse{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})

}

func (apiCfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type ValidResponse struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "parsing error", err)
	}
	user, err := apiCfg.db.GetUserByEmail(r.Context(), params.Email)

	pwMatch, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)

	if !pwMatch {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	expiryTime := 3600

	token, err := auth.MakeJWT(user.ID, apiCfg.jwtSecret, time.Duration(expiryTime)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't create token", err)
		return
	}

	rToken := auth.MakeRefreshToken()
	rTokenExpiry := time.Now().Add(60 * 24 * time.Hour)

	apiCfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: rToken, UserID: user.ID, ExpiresAt: rTokenExpiry})

	respondWithJSON(w, http.StatusOK, ValidResponse{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: rToken,
	})
}

func (apiCfg *apiConfig) createUsersHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type ValidResponse struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "parsing error", err)
	}

	hashedPassword, err := auth.HashPassword(params.Password)

	user, err := apiCfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hashedPassword})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, ValidResponse{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})

}
