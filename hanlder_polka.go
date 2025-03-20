package main

import (
	"encoding/json"
	"net/http"

	"github.com/burush0/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolka(w http.ResponseWriter, r *http.Request) {
	api_key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", err)
		return
	}

	if api_key != cfg.polka_key {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", nil)
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.db.UpdateUserChirpyRed(r.Context(), params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
