package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/disconnectedag/go-server/internal/auth"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) upgradeToRedHandler(w http.ResponseWriter, r *http.Request) {
	type EventData struct {
		UserID string `json:"user_id"`
	}
	type parameters struct {
		Event string    `json:"event"`
		Data  EventData `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	log.Print(apiKey)

	if apiCfg.polkaKey != apiKey {
		w.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Parsing error", err)
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	userUUID, err := uuid.Parse(params.Data.UserID)

	_, err = apiCfg.db.UpgradeUser(r.Context(), userUUID)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(204)
}
