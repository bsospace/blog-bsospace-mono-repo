package media

import (
	"mime/multipart"

	"github.com/google/uuid"
)

type MediaResponse struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	ImageURL string    `json:"image_url"`
	IsUsed   bool      `json:"is_used"`
}

type UploadMediaForm struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}
