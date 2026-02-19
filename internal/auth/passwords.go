package auth

import (
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"errors"
	"github.com/google/uuid"
	"strings"
	"net/http"
	"crypto/rand"
	"encoding/hex"
)

type Claims struct{
	Role		string		`json:"role"`
	jwt.RegisteredClaims
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
		return match, err
	}

	return match, nil
}

func MakeJWT(userID uuid.UUID, role, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := &Claims{
		Role:				role,
		RegisteredClaims:	jwt.RegisteredClaims{
			Issuer: 	"spellbook-access",
			IssuedAt: 	jwt.NewNumericDate(now),
			ExpiresAt: 	jwt.NewNumericDate(now.Add(expiresIn)),
			Subject: 	userID.String(),
		},
	}

	 token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	 signingKey := []byte(tokenSecret)

	 signedToken, err := token.SignedString(signingKey)
	 if err != nil {
		return "", err
	 }

	 return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, string, error) {
    claims := &Claims{}

    _, err := jwt.ParseWithClaims(
        tokenString,
        claims,
        func(token *jwt.Token) (interface{}, error) {
            return []byte(tokenSecret), nil
        },
    )
    if err != nil {
        return uuid.Nil, "", err
    }

    issuer, err := claims.GetIssuer()
    if err != nil {
        return uuid.Nil, "", err
    }
    if issuer != "spellbook-access" {
        return uuid.Nil, "", errors.New("invalid issuer")
    }

    userIDString, err := claims.GetSubject()
    if err != nil {
        return uuid.Nil, "", err
    }

    id, err := uuid.Parse(userIDString)
    if err != nil {
        return uuid.Nil, "", err
    }

    return id, claims.Role, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")

	if authHeader == "" {
		return "", errors.New("Authorization header missing")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("Authorization has improper syntax")
	}

	sliced := strings.TrimPrefix(authHeader, "Bearer ")
	sliced = strings.TrimSpace(sliced)

	if sliced == "" {
		return "", errors.New("Authorization missing token string")
	}

	return sliced, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	n, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	if n != len(key) {
		return "", errors.New("Error generating refresh token")
	}

	refreshToken := hex.EncodeToString(key)

	return refreshToken, nil
}
