package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func JsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Hash(origin string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(origin), bcrypt.DefaultCost)
	if string(hash) == "" || err != nil {
		fmt.Printf("Error encrypting password hash")
		return ""
	}

	return string(hash)
}

func CheckHash(hash, password []byte) bool {
	if err := bcrypt.CompareHashAndPassword(password, hash); err != nil {
		return false
	}

	return true
}

func GenerateAccessToken(userID uint) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if err := godotenv.Load(); err != nil {
		return ""
	}
	tokenString, err := token.SignedString([]byte(os.Getenv("SIGN_SECRET")))
	if err != nil {
		return ""
	}

	return tokenString
}

func GenerateRefreshToken() string {
	return rand.Text()
}
