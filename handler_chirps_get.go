package main

import (
	"log"
	"net/http"

	"github.com/disconnectedag/go-server/internal/database"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := apiCfg.db.GetAllChirps(r.Context())
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

func dbChirpToChirp(c database.Chirp) Chirp {
	return Chirp{
		ID:        c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID,
	}
}

