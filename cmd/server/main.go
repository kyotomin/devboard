package main

import (
	"context"
	"devboard/internal/auth"
	"devboard/internal/boards"
	"devboard/internal/middleware"
	"devboard/internal/storage"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"

	"gorm.io/gorm"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := storage.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к бд: %v", err)
	}

	r2, err := storage.InitR2()
	if err != nil {
		log.Fatal("Error while loading r2 storage")
	}

	mux := NewRouter(db, r2)

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Ошибка получения sql.DB: %v", err)
	}
	defer func() {
		fmt.Println("Закрытие пула соединений БД...")
		sqlDB.Close()
	}()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("Сервер запущен на %v", os.Getenv("PORT"))
		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Ошибка сервера.")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Ошибка при остановке сервера: %v", err)
	}
	fmt.Println("Сервер остановлен")
}

func NewRouter(db *gorm.DB, r2 *s3.Client) http.Handler {
	mux := http.NewServeMux()

	// Создание хендлеров с инъекцией БД
	authHandler := auth.NewHandler(db)
	boardHandler := boards.NewHandler(db, r2)

	// Регистрация хендлеров авторизации
	mux.HandleFunc("POST /api/auth/register", authHandler.HandleRegistration)
	mux.HandleFunc("POST /api/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("POST /api/auth/refresh", authHandler.HandleRefresh)

	// Регистрация хендлеров доски
	mux.HandleFunc("POST /api/boards/", middleware.ValidateCookie(boardHandler.HandleCreateBoard))
	mux.HandleFunc("GET /api/boards/{boardID}", middleware.ValidateCookie(boardHandler.HandleGetBoard))
	mux.HandleFunc("POST /api/boards/{boardID}/columns", middleware.ValidateCookie(boardHandler.HandleCreateColunm))
	mux.HandleFunc("POST /api/boards/{boardID}/columns/{columnID}/cards", middleware.ValidateCookie(boardHandler.HandleCreateCard))
	mux.HandleFunc("POST /api/boards/{boardID}/columns/{columnID}/cards/{cardID}/attachments", middleware.ValidateCookie(boardHandler.HandleUploadAttachment))

	return mux
}
