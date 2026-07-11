package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenCookieName = "token"
	tokenExpiry     = 8 * time.Hour
)

type signinRequest struct {
	Password string `json:"password"`
}

type signinResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

func getPasswordHash(password string) string {
	if password == "" {
		return ""
	}
	return fmt.Sprintf("%x", password)
}

func generateToken(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("Пароль не может быть пустым")
	}
	passwordHash := getPasswordHash(password)

	claims := Claims{
		PasswordHash: passwordHash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := []byte(password)

	return token.SignedString(secret)
}

func validateToken(tokenString string, currentPassword string) bool {
	if tokenString == "" || currentPassword == "" {
		return false
	}

	claims := &Claims{}
	secret := []byte(currentPassword)

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) { return secret, nil })

	if err != nil {
		log.Println("Ошибка токена: %w", err)
		return false
	}

	if !token.Valid {
		log.Println("Токен не валиден")
		return false
	}

	expectedHash := getPasswordHash(currentPassword)
	if claims.PasswordHash != expectedHash {
		log.Printf("Хэш пароля не совпадает: ожидается %s, получено %s", expectedHash, claims.PasswordHash)
		return false
	}

	return true
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(signinResponse{Error: "Пароль не настроен"})
		return
	}

	expectedPassword := os.Getenv("TODO_PASSWORD")
	if expectedPassword == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(signinResponse{Error: "Пароль не настроен"})
		return
	}

	var req signinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(signinResponse{Error: "Неверный формат JSON"})
		return
	}

	if req.Password != expectedPassword {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(signinResponse{Error: "Пароль неверный"})
		return
	}

	token, err := generateToken(req.Password)
	if err != nil {
		log.Printf("Ошибка генерации токена: %v", err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(signinResponse{Error: "Ошибка генерации токена"})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(signinResponse{Token: token})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		password := os.Getenv("TODO_PASSWORD")

		if password == "" {
			next(w, r)
			return
		}

		cookie, err := r.Cookie(tokenCookieName)

		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		if !validateToken(cookie.Value, password) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}

		next(w, r)
	}
}
