package main

import (
	"context"
	"devboard/internal/auth"
	"devboard/internal/storage"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	db, err := storage.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к бд: %v", err)
	}

	mux := NewRouter(db)

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

func NewRouter(db *gorm.DB) http.Handler {
	mux := http.NewServeMux()

	// Создание хендлеров с инъекцией БД
	authHandler := auth.NewHandler(db)

	// Регистрация хендлеров
	mux.HandleFunc("/api/auth/register", authHandler.HandleRegistration)

	return mux
}
