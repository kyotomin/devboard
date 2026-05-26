package boards

import (
	"context"
	"devboard/internal/utils"
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) RequireRole(roles ...ContributorRole) func(http.HandlerFunc) http.HandlerFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("userID").(uuid.UUID)
			boardID := r.PathValue("boardID")

			contributor := &BoardContributor{}
			err := h.db.Where("board_id = ? and user_id = ?", boardID, userID).First(&contributor).Error
			if err != nil {
				utils.JsonResponse(w, http.StatusForbidden, "Недостаточно прав для добавления")
			}

			for _, role := range roles {
				if contributor.Role == role {
					ctx := context.WithValue(r.Context(), "contributor", contributor)
					hf.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			utils.JsonResponse(w, 403, "Недостаточно прав")
		}
	}
}
