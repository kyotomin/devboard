package auth

import (
	"crypto/sha256"
	"devboard/internal/utils"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	req := RegisterRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(w, 400, "Error occured, try again later")
		return
	}

	if req.Email == "" || req.Username == "" || req.Password == "" {
		http.Error(w, "Заполните все поля ввода", 401)
		return
	}

	user := User{
		Username:  req.Username,
		Email:     req.Email,
		Hash:      utils.Hash(req.Password),
		CreatedAt: time.Now(),
	}

	if res := h.db.Create(&user); res.Error != nil {
		if strings.Contains(res.Error.Error(), "duplicate key") {
			http.Error(w, "Username или Email уже существует", http.StatusUnauthorized)
			return
		}

		http.Error(w, "Ошибка создания пользователя", http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, 200, "User created")
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	req := LoginRequest{}
	user := User{}
	json.NewDecoder(r.Body).Decode(&req)
	if err := h.db.Where("email = ?", req.Email).First(&user); err.Error != nil {
		if err.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Неверный email", http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, "Ошибка с авторизацией", http.StatusUnauthorized)
			return
		}
	}

	if !utils.CheckHash([]byte(user.Hash), []byte(req.Password)) {
		http.Error(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	accessToken := utils.GenerateAccessToken(user.ID)
	refreshToken := utils.GenerateRefreshToken()
	hash := sha256.Sum256([]byte(refreshToken))

	t := h.db.Create(&RefreshToken{
		UserID:    user.ID,
		TokenHash: hex.EncodeToString(hash[:]),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})

	if t.Error != nil {
		http.Error(w, "Ошибка создания токена авторизации", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   900,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   30 * 24 * 60 * 60,
		Path:     "/api/auth/refresh",
	})

	utils.JsonResponse(w, 200, "Успешная авторизация")
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh token either revoked or does not exists", 401)
		return
	}

	sum := sha256.Sum256([]byte(cookie.Value))
	hash := hex.EncodeToString(sum[:])

	refreshToken := &RefreshToken{}
	result := h.db.Where("token_hash = ? AND expires_at > ?", hash, time.Now()).First(&refreshToken)
	if result.Error != nil {
		http.Error(w, "Refresh не валиден", 401)
		return
	}

	accessToken := utils.GenerateAccessToken(refreshToken.UserID)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   900,
		Path:     "/",
	})

	utils.JsonResponse(w, 200, "Токен обновлен")
}
