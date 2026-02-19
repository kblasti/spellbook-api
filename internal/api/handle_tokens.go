package api

import (
	"github.com/kblasti/spellbook/internal/auth"
	"net/http"
	"time"
)

const expirationTime = time.Duration(3600) * time.Second

func (cfg *APIConfig) HandlerRefresh(w http.ResponseWriter, r *http.Request) {
    type tokenResponse struct {
        Token string `json:"token"`
    }
    
    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, 400, "Error with bearer token")
        return
    }

    user, err := cfg.DB.GetUserFromRefreshToken(r.Context(), token)
    if err != nil {
        respondWithError(w, 401, "Error finding user")
        return
    }

    jwtToken, err := auth.MakeJWT(user.ID, user.Role, cfg.Secret, expirationTime)
    if err != nil {
        respondWithError(w, 500, "Error making token")
        return
    }

    response := tokenResponse {
        Token: jwtToken,
    }

    respondWithJSON(w, 200, response)
    return
}

func (cfg *APIConfig) HandlerRevoke(w http.ResponseWriter, r *http.Request) {
    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, 400, "Error with bearer token")
        return
    }

    err = cfg.DB.RevokeRefreshToken(r.Context(), token)
    if err != nil {
        respondWithError(w, 500, "Error revoking token")
        return
    }

    respondWithJSON(w, 204, nil)
    return
}