package auth

import (
	"devboard/internal/utils"
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
	if r.Method != "POST" {
		utils.JsonResponse(w, http.StatusMethodNotAllowed, "Unavailable method")
		return
	}

	req := RegisterRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(w, 400, "Error occured, try again later")
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

func (h *Handler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.JsonResponse(w, http.StatusMethodNotAllowed, "Unavailable method")
		return
	}

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

	t := h.db.Create(&RefreshToken{
		UserID:    user.ID,
		TokenHash: utils.Hash(refreshToken),
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
		Path:     "/auth/refresh",
	})

	utils.JsonResponse(w, 200, "Успешная авторизация")
}
