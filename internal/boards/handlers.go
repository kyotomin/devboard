package boards

import (
	"context"
	"devboard/internal/auth"
	"devboard/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
	r2 *s3.Client
}

func NewHandler(db *gorm.DB, r2 *s3.Client) *Handler {
	return &Handler{db: db, r2: r2}
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

func (h *Handler) HandleCreateCard(w http.ResponseWriter, r *http.Request) {
	columnID := r.PathValue("columnID")
	boardID := r.PathValue("boardID")
	asigneeID := r.Context().Value("userID").(uuid.UUID)

	column := &Column{}
	if err := h.db.Where("id = ? and board_id = ?", columnID, boardID).First(&column).Error; err != nil {
		utils.JsonResponse(w, 404, "Колонка не найдена")
		return
	}

	req := CreateCardRequest{}
	json.NewDecoder(r.Body).Decode(&req)

	var maxPos int
	h.db.Model(&Card{}).
		Where("column_id = ?", boardID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPos)

	card := &Card{
		ColumnID:    column.ID,
		Title:       req.Title,
		Attachments: make([]Attachment, 0),
		CreatedAt:   time.Now(),
		AssigneeID:  &asigneeID,
		Position:    maxPos,
	}

	if err := h.db.Create(&card).Error; err != nil {
		utils.JsonResponse(w, 500, "Ошибка создания карточки")
		return
	}

	utils.JsonResponse(w, 200, "Карточка создана")
}

func (h *Handler) UploadNewAttachment(file multipart.File, header *multipart.FileHeader, cardID uuid.UUID) (*Attachment, error) {
	key := fmt.Sprintf("attachments/%s/%s", cardID, header.Filename)

	_, err := h.r2.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(os.Getenv("R2_BUCKET")),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(header.Size),
		ContentType:   aws.String(header.Header.Get("Content-Type")),
	})

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", os.Getenv("R2_PUBLIC_URL"), key)

	return &Attachment{
		CardID:    cardID,
		FileURL:   url,
		FileName:  header.Filename,
		CreatedAt: time.Now(),
	}, nil
}

func (h *Handler) HandleUploadAttachment(w http.ResponseWriter, r *http.Request) {
	cardId := r.PathValue("cardID")

	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		utils.JsonResponse(w, 500, "Ошибка получения файла")
		return
	}

	attachment, err := h.UploadNewAttachment(file, header, uuid.MustParse(cardId))
	if err != nil {
		log.Printf("R2 upload error: %v", err)
		utils.JsonResponse(w, 500, "Ошибка загрузки файла")
		return
	}

	if err := h.db.Create(&attachment).Error; err != nil {
		utils.JsonResponse(w, 500, "Ошибка сохранения")
		return
	}

	utils.JsonResponse(w, 201, "Файл загружен")
}

func (h *Handler) HandleGetBoard(w http.ResponseWriter, r *http.Request) {
	boardID := r.PathValue("boardID")

	board := &Board{}
	if err := h.db.
		Preload("Columns", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Columns.Cards", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Columns.Cards.Attachments").
		First(&board, "id = ?", boardID).Error; err != nil {
		utils.JsonResponse(w, 404, "Борда не найдена")
		return
	}

	json.NewEncoder(w).Encode(board)
}

func (h *Handler) HandleAddContributor(w http.ResponseWriter, r *http.Request) {
	boardID := r.PathValue("boardID")

	req := &AddContributorRequest{}
	json.NewDecoder(r.Body).Decode(&req)

	user := &auth.User{}
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		utils.JsonResponse(w, 404, "Юзер не найден")
		return
	}

	contributor := &BoardContributor{
		BoardID: uuid.MustParse(boardID),
		UserID:  user.ID,
		Role:    req.Role,
	}

	if err := h.db.Create(&contributor).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			utils.JsonResponse(w, 409, "Юзер уже в борде")
			return
		}
		utils.JsonResponse(w, 500, "Ошибка добавления")
		return
	}

	utils.JsonResponse(w, 200, "Контрибьютор добавлен")
}

func (h *Handler) HandleDeleteContributor(w http.ResponseWriter, r http.Request) {
	boardID := r.PathValue("boardID")

	req := &DeleteContributorRequest{}
	json.NewDecoder(r.Body).Decode(&req)

	res := h.db.Delete(&BoardContributor{}, "board_id = ? and user_id = ?",
		uuid.MustParse(boardID), uuid.MustParse(req.UserID.String()))

	if res.Error != nil {
		utils.JsonResponse(w, 500, "Ошибка удаления")
	}

	if res.RowsAffected == 0 {
		utils.JsonResponse(w, 404, "Контрибьютор не найден")
		return
	}

	utils.JsonResponse(w, 200, "Контрибьютор удалён")
}

func (h *Handler) HandleGetAllContributors(w http.ResponseWriter, r *http.Request) {
	boardID := r.PathValue("boardID")

	var contributors []BoardContributor
	res := h.db.Preload("board_id = ?", uuid.MustParse(boardID)).Find(&contributors)

	if res.Error != nil {
		utils.JsonResponse(w, 500, "Ошибка получения")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(contributors)
}

func (h *Handler) HandleMoveCard(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("cardID")
	req := &MoveCardRequest{}
	json.NewDecoder(r.Body).Decode(&req)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		tx.Model(&Card{}).Where("column_id = ? and position = ?", req.ColumnID, req.Position).
			UpdateColumn("position", gorm.Expr("position + 1"))

		return tx.Model(&Card{}).Where("id = ?", cardID).
			Updates(map[string]any{
				"column_id": req.ColumnID,
				"position":  req.Position,
			}).Error
	})

	if err != nil {
		utils.JsonResponse(w, 500, "Ошибка перемещения")
		return
	}

	utils.JsonResponse(w, 200, "Карточка перемещена")
}
