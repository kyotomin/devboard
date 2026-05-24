package boards

import (
	"devboard/internal/auth"
	"time"

	"github.com/google/uuid"
)

type ContributorRole string

const (
	RoleOwner   ContributorRole = "owner"
	RoleManager ContributorRole = "manager"
	RoleWorker  ContributorRole = "worker"
)

type BoardContributor struct {
	BoardID uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID  uuid.UUID       `gorm:"type:uuid;primaryKey"`
	User    auth.User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Role    ContributorRole `gorm:"type:varchar(20);default:'worker';not null"`
}

type Column struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	BoardID  uuid.UUID `gorm:"type:uuid;not null"`
	Name     string    `gorm:"not null"`
	Cards    []Card    `gorm:"foreignKey:ColumnID"`
	Position int       `gorm:"default:0"`
}

type Card struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ColumnID    uuid.UUID
	Title       string
	Attachments []Attachment
	CreatedAt   time.Time
	UpdatedAt   time.Time
	AssigneeID  *uuid.UUID `gorm:"type:uuid"`
	Position    int        `gorm:"default:0"`
}

type Board struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OwnerID      uuid.UUID
	Title        string `gorm:"unique;not null"`
	Description  string
	Contributors []BoardContributor `gorm:"foreignKey:BoardID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateBoardRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateColumnRequest struct {
	Name string `json:"name"`
}

type CreateCardRequest struct {
	Title string `json:"title"`
}

type Attachment struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CardID    uuid.UUID `gorm:"type:uuid;not null;index"` // Индекс для быстрого поиска
	FileURL   string    `gorm:"type:text;not null"`       // Ссылка на CDN
	FileName  string    `gorm:"type:varchar(255)"`        // Оригинальное имя файла
	CreatedAt time.Time
}
