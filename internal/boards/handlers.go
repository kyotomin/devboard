package boards

import (
	"devboard/internal/utils"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) HandleCreateBoard(w http.ResponseWriter, r *http.Request) {
	ownerID := r.Context().Value("userID").(uuid.UUID)

	req := &CreateBoardRequest{}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Title == "" {
		utils.JsonResponse(w, 400, "Title обязателен")
		return
	}

	board := &Board{
		OwnerID:      ownerID,
		Title:        req.Title,
		Description:  req.Description,
		Contributors: make([]BoardContributor, 0),
		CreatedAt:    time.Now(),
	}

	if err := h.db.Create(&board).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			utils.JsonResponse(w, 409, "Борда с таким названием уже существует")
			return
		}
		utils.JsonResponse(w, 500, "Ошибка создания борды")
		return
	}
}

func (h *Handler) HandleCreateColunm(w http.ResponseWriter, r *http.Request) {
	boardID := r.PathValue("boardID")
	req := CreateColumnRequest{}
	json.NewDecoder(r.Body).Decode(&req)

	var maxPos int
	h.db.Model(&Column{}).
		Where("board_id = ?", boardID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPos)

	column := &Column{
		BoardID:  uuid.MustParse(boardID),
		Name:     req.Name,
		Cards:    make([]Card, 0),
		Position: maxPos,
	}

	if err := h.db.Create(&column).Error; err != nil {
		utils.JsonResponse(w, 500, "Ошибка создания колонки")
		return
	}

	utils.JsonResponse(w, 200, "Колонка создана")
}
