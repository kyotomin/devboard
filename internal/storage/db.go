package storage

import (
	"context"
	"devboard/internal/auth"
	"devboard/internal/boards"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(
		&auth.User{},
		&auth.RefreshToken{},
		&boards.Board{},
		&boards.BoardContributor{},
		&boards.Column{},
		&boards.Card{},
		&boards.Attachment{},
	)

	if err != nil {
		log.Fatal("Ошибка создания таблицы пользователей")
	}

	return db, err
}

func InitR2() (*s3.Client, error) {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY")
	secretKey := os.Getenv("R2_SECRET_KEY")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey, secretKey, "",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
	})

	return client, nil
}
