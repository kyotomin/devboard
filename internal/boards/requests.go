package boards

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
