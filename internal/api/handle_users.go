package api

import (
	"github.com/kblasti/spellbook/internal/database"
	"github.com/kblasti/spellbook/internal/auth"
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"time"
	"net/http"
	"strings"
	"context"
)

type User struct {
	ID        	uuid.UUID 	`json:"id"`
	CreatedAt 	time.Time 	`json:"created_at"`
	UpdatedAt 	time.Time 	`json:"updated_at"`
	Email     	string    	`json:"email"`
	Role		string		`json:"role"`
}

func (cfg *APIConfig) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
    type Input struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    role := "user"

    decoder := json.NewDecoder(r.Body)
    input := Input{}

    err := decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding post")
        return
    }

    if !IsValidEmailFormat(input.Email) {
        respondWithError(w, 400, "Invalid email format")
        return
    }

    if IsDisposableEmail(input.Email) {
        respondWithError(w, 400, "Disposable email addresses are not allowed")
        return
    }

    if len(input.Password) < 6 {
        respondWithError(w, 400, "Password must be at least 6 characters")
        return
    }

    hashed, err := auth.HashPassword(input.Password)
    if err != nil {
        respondWithError(w, 500, "Error hashing password")
        return
    }

    dbUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
        Email:          input.Email,
        HashedPassword: hashed,
        Role:           role,
    })
    if err != nil {
        respondWithError(w, 500, "Error creating user")
        return
    }

    appUser := User{
        ID:        dbUser.ID,
        CreatedAt: dbUser.CreatedAt,
        UpdatedAt: dbUser.UpdatedAt,
        Email:     dbUser.Email,
        Role:      dbUser.Role,
    }

    respondWithJSON(w, 201, appUser)
}

func (cfg *APIConfig) HandlerCreateAdminUser(w http.ResponseWriter, r *http.Request) {
	type Input struct {
        Email string `json:"email"`
        Password string `json:"password"`
    }

    role := "admin"

    decoder := json.NewDecoder(r.Body)
    input := Input{}

    err := decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding post")
        return
    }

    hashed, err := auth.HashPassword(input.Password)
    if err != nil {
        respondWithError(w, 500, "Error hashing password")
        return
    }

    dbUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
        Email:          input.Email,
        HashedPassword: hashed,
		Role:			role,
    })
    if err != nil {
        respondWithError(w, 500, "Error creating user")
        return
    }

    appUser := User{
        ID:         dbUser.ID,
        CreatedAt:  dbUser.CreatedAt,
        UpdatedAt:  dbUser.UpdatedAt,
        Email:      dbUser.Email,
		Role:		dbUser.Role,
    }

    respondWithJSON(w, 201, appUser)
    return
}

func (cfg *APIConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
    type Input struct {
        Email string `json:"email"`
        Password string `json:"password"`
    }

    type loginResponse struct {
        User
        Token string `json:"token"`
        RefreshToken string `json:"refresh_token"`
    }

    decoder := json.NewDecoder(r.Body)
    input := Input{}

    err := decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Something went wrong")
        return
    }

    dbUser, err := cfg.DB.UserLogin(r.Context(), input.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            respondWithError(w, 401, "Incorrect email or password")
            return
        } else {
            respondWithError(w, 500, "Something went wrong")
            return
        }
    }

    verified, err := auth.CheckPasswordHash(input.Password, dbUser.HashedPassword)
    if err != nil {
        respondWithError(w, 500, "Something went wrong")
        return
    }

    if verified == false {
        respondWithError(w, 401, "Incorrect email or password")
        return
    }

    token, err := auth.MakeJWT(dbUser.ID, dbUser.Role, cfg.Secret, expirationTime)
    if err != nil {
        respondWithError(w, 500, "Error making token")
        return
    }

    refreshToken, err := auth.MakeRefreshToken()
    if err != nil {
        respondWithError(w, 500, "Error making refresh token")
        return
    }

    dbUserID := uuid.NullUUID{
        UUID:   dbUser.ID,
        Valid:  true,
    }

    _, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
        Token: refreshToken,
        UserID: dbUserID,
    })
    if err != nil {
        respondWithError(w, 500, "Error saving refresh token")
        return
    }

    appUser := User{
        ID:         dbUser.ID,
        CreatedAt:  dbUser.CreatedAt,
        UpdatedAt:  dbUser.UpdatedAt,
        Email:      dbUser.Email,
		Role:		dbUser.Role,
    }

    login := loginResponse{
        User:  appUser,
        Token: token,
        RefreshToken: refreshToken,
    }

    respondWithJSON(w, 200, login)
    return
}

func (cfg *APIConfig) HandlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		Password		string			`json:"password"`
	}

	token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, 401, "Error retrieving token")
        return
    }

    userID, _, err := auth.ValidateJWT(token, cfg.Secret)
    if err != nil {
        respondWithError(w, 401, "Error validating token")
        return
    }

	decoder := json.NewDecoder(r.Body)
    input := Input{}

    err = decoder.Decode(&input)
    if err != nil {
        respondWithError(w, 500, "Error decoding input")
        return
    }

    dbUser, err := cfg.DB.GetHashedPassword(r.Context(), userID)
    if err != nil {
        respondWithError(w, 500, "Unable to retrieve user data")
        return
    }

	verified, err := auth.CheckPasswordHash(input.Password, dbUser.HashedPassword)
    if err != nil {
        respondWithError(w, 500, "Error verifying password")
        return
    }

    if verified == false {
        respondWithError(w, 401, "Incorrect password")
        return
    }

    err = cfg.DB.DeleteUser(r.Context(), dbUser.ID)
    if err != nil {
        respondWithError(w, 500, "Error deleting user")
        return
    }

	respondWithMessage(w, 201, "User deleted")
	return
}

func (cfg *APIConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
    type parameters struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, 401, "Error retrieving token")
        return
    }

    userID, _, err := auth.ValidateJWT(token, cfg.Secret)
    if err != nil {
        respondWithError(w, 401, "Error validating token")
        return
    }

    decoder := json.NewDecoder(r.Body)
    params := parameters{}

    err = decoder.Decode(&params)
    if err != nil {
        respondWithError(w, 500, "Error decoding input")
        return
    }

    if params.Email != "" {
        if !IsValidEmailFormat(params.Email) {
            respondWithError(w, 400, "Invalid email format")
            return
        }

        if IsDisposableEmail(params.Email) {
            respondWithError(w, 400, "Disposable email addresses are not allowed")
            return
        }
    }

    var hashedPassword string
    if params.Password != "" {
        if len(params.Password) < 6 {
            respondWithError(w, 400, "Password must be at least 6 characters")
            return
        }

        hashedPassword, err = auth.HashPassword(params.Password)
        if err != nil {
            respondWithError(w, 500, "Error hashing password")
            return
        }
    }

    currentUser, err := cfg.DB.GetUserByID(r.Context(), userID)
    if err != nil {
        respondWithError(w, 500, "Error retrieving user")
        return
    }

    emailToSave := currentUser.Email
    if params.Email != "" {
        emailToSave = params.Email
    }

    passwordToSave := currentUser.HashedPassword
    if hashedPassword != "" {
        passwordToSave = hashedPassword
    }

    dbUser, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
        Email:          emailToSave,
        HashedPassword: passwordToSave,
        Role:           "user",
        ID:             userID,
    })
    if err != nil {
        respondWithError(w, 500, "Error updating user")
        return
    }

    response := User{
        ID:        dbUser.ID,
        CreatedAt: dbUser.CreatedAt,
        UpdatedAt: dbUser.UpdatedAt,
        Email:     dbUser.Email,
    }

    respondWithJSON(w, 200, response)
}

func (cfg *APIConfig) AuthMiddleware(next http.Handler) http.Handler { 
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { 
		authHeader := r.Header.Get("Authorization") 
		if authHeader == "" { 
			http.Error(w, "missing authorization header", http.StatusUnauthorized) 
			return 
		} 
		
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ") 
		if tokenStr == authHeader { 
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized) 
			return 
		} 
		
		userID, _, err := auth.ValidateJWT(tokenStr, cfg.Secret) 
		if err != nil { 
			http.Error(w, "invalid token", http.StatusUnauthorized) 
			return 
		} 
		
		ctx := context.WithValue(r.Context(), "userID", userID) 
		next.ServeHTTP(w, r.WithContext(ctx)) 
		}) 
	}
func (cfg *APIConfig) AdminOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims, ok := r.Context().Value("claims").(*auth.Claims)
        if !ok || claims.Role != "admin" {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
