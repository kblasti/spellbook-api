package api

import (
	"net/http"
	"encoding/json"
	"github.com/kblasti/spellbook/internal/database"
)

type APIConfig struct {
  DB        *database.Queries
  Platform  string
  Secret    string
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    payload := map[string]string{
        "error": msg,
    }
    data, err := json.Marshal(payload)
    if err != nil {
        http.Error(w, "Something went wrong", http.StatusInternalServerError)
        return
    }

    w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    data, err := json.Marshal(payload)
    if err != nil {
        respondWithError(w, 500, "Something went wrong")
        return
    }

    w.Write(data)
}

func respondWithMessage(w http.ResponseWriter, code int, msg string) {
    respondWithJSON(w, code, map[string]string{
        "message": msg,
    })
}


func EnableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

        if r.Method == "OPTIONS" {
            return
        }

        next.ServeHTTP(w, r)
    })
}

