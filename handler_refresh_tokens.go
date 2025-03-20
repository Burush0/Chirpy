package main

import (
	"net/http"
	"time"

	"github.com/burush0/chirpy/internal/auth"
)

type AccessToken struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", err)
		return
	}

	if time.Until(refreshToken.ExpiresAt) <= 0 {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", err)
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", err)
		return
	}

	user, err := cfg.db.GetUserByRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user by refresh token", err)
		return
	}

	expiration := 1 * time.Hour

	tokenString, err := auth.MakeJWT(user.ID, cfg.jwt_secret, expiration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, AccessToken{tokenString})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
