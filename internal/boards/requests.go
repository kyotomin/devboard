package boards

import "github.com/google/uuid"

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

type AddContributorRequest struct {
	Username string          `json:"type:varchar(255);not null"`
	Role     ContributorRole `json:"role"`
}

type DeleteContributorRequest struct {
	UserID uuid.UUID `json:"userID"`
}

type MoveCardRequest struct {
	ColumnID uuid.UUID `json:"column_id"`
	Position int       `json:"position"`
}

type GetBoardTitle struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}
