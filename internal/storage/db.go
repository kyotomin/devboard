package storage

import (
	"devboard/internal/auth"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&auth.User{})
	if err != nil {
		log.Fatal("Ошибка создания таблицы пользователей")
	}

	return db, err
}
