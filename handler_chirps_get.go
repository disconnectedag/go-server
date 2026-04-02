package main

import (
	"database/sql"
	"log"
	"net/http"
	"sort"

	"github.com/disconnectedag/go-server/internal/auth"
	"github.com/disconnectedag/go-server/internal/database"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")
	var chirps []database.Chirp
    var err error
	if authorId != "" {
		authorId, _ := uuid.Parse(authorId)
		chirps, err = apiCfg.db.GetAllChirpsByUserID(r.Context(), authorId)
	} else {
		chirps, err = apiCfg.db.GetAllChirps(r.Context())
	}
	if sortOrder != "" {
		if sortOrder == "asc" {
			sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)})
		}
		if sortOrder == "desc" {
			sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt)})
		}
	}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	response := []Chirp{}
	for _, c := range chirps {
		response = append(response, dbChirpToChirp(c))
	}
	respondWithJSON(w, http.StatusOK, response)

}
func (apiCfg *apiConfig) getOneChirpHandler(w http.ResponseWriter, r *http.Request) {
	type ValidResponse struct {
		Chirp
	}
	id := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(id)
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
		respondWithError(w, http.StatusNotFound, err.Error(), err)
		return
	}
	chirp, err := apiCfg.db.GetOneChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error(), err)
		return
	}
	respondWithJSON(w, http.StatusOK, ValidResponse{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		},
	})
}
func (apiCfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(id)
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
		respondWithError(w, http.StatusNotFound, err.Error(), err)
		return
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

	chirp, err := apiCfg.db.GetOneChirp(r.Context(), chirpID)

	if err == sql.ErrNoRows {
		w.WriteHeader(404)
		return
	}

	if err != nil {
		respondWithError(w, 500, "Couldn't fetch chirp", err)
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "unauthorized to delete this chirp", err)
		return
	}
	apiCfg.db.DeleteChirp(r.Context(), chirpID)
	w.WriteHeader(204)
}

func dbChirpToChirp(c database.Chirp) Chirp {
	return Chirp{
		ID:        c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID,
	}
}
