package media

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MediaServiceInterface interface {
	CreateMedia(fileHeader *multipart.FileHeader, user *models.User, postID *uuid.UUID) (*models.ImageUpload, error)
	DeleteFromChibisafe(image *models.ImageUpload) error
	GetImagesByPostID(postID uuid.UUID) ([]models.ImageUpload, error)
	UpdateImageUsage(image *models.ImageUpload) error
	DeleteUnusedImages() error
	GetImageByURL(imageURL string) (*models.ImageUpload, error)
	UploadToChibisafe(fileHeader *multipart.FileHeader) (ChibisafeResponse, error)
}

type MediaService struct {
	Repo   MediaRepositoryInterface
	Logger *zap.Logger
}

func NewMediaService(repo MediaRepositoryInterface, logger *zap.Logger) MediaServiceInterface {
	return &MediaService{Repo: repo, Logger: logger}
}

func (s *MediaService) CreateMedia(fileHeader *multipart.FileHeader, user *models.User, postID *uuid.UUID) (*models.ImageUpload, error) {

	// Upload ไปยัง Chibisafe (สมมุติ)
	res, err := s.UploadToChibisafe(fileHeader)
	if err != nil {
		return nil, err
	}

	s.Logger.Info("RESULT - Chibisafe Response Name:", zap.String("name", res.Name))
	s.Logger.Info("RESULT - Chibisafe Response Identifier:", zap.String("identifier", res.Identifier))
	s.Logger.Info("RESULT - Chibisafe Response:", zap.Any("response", res))
	s.Logger.Info("RESULT - Chibisafe Response UUID:", zap.String("uuid", res.UUID))
	// log user
	s.Logger.Info("RESULT - User ID:", zap.String("user_id", user.ID.String()))
	s.Logger.Info("RESULT - User Email:", zap.String("user_email", user.Email))

	image := &models.ImageUpload{
		ID:         uuid.New(),
		ImageURL:   res.URL,
		IsUsed:     false,
		UserID:     user.ID,
		PostID:     postID,
		UsedReason: "Blog image",
		FileID:     res.UUID,
	}

	if err := s.Repo.Create(image); err != nil {
		return nil, err
	}

	return image, nil
}

type ChibisafeResponse struct {
	Name       string `json:"name"`
	UUID       string `json:"uuid"`
	URL        string `json:"url"`
	Identifier string `json:"identifier"`
	PublicURL  string `json:"public_url"`
}

func addRandomSuffixToFile(file multipart.File) ([]byte, error) {
	// อ่านไฟล์ทั้งหมดเข้าหน่วยความจำ
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// สร้าง random byte (8 bytes)
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}

	// ต่อท้ายไฟล์ด้วย randomBytes
	data = append(data, randomBytes...)
	return data, nil
}

func (s *MediaService) UploadToChibisafe(fileHeader *multipart.FileHeader) (ChibisafeResponse, error) {
	cfg := config.LoadConfig()
	chibisafeURL := cfg.ChibisafeURL
	chibisafeToken := cfg.ChibisafeKey
	albmnId := cfg.ChibisafeAlbumId

	file, err := fileHeader.Open()
	if err != nil {
		return ChibisafeResponse{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// เปลี่ยนไฟล์ให้ checksum เปลี่ยน
	modifiedData, err := addRandomSuffixToFile(file)
	if err != nil {
		return ChibisafeResponse{}, fmt.Errorf("failed to modify file: %w", err)
	}

	// Prepare multipart/form-data body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileHeader.Filename+"-"+uuid.New().String())
	if err != nil {
		return ChibisafeResponse{}, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(modifiedData); err != nil {
		return ChibisafeResponse{}, fmt.Errorf("failed to copy file: %w", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", chibisafeURL+"/api/upload", body)
	if err != nil {
		return ChibisafeResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-api-key", chibisafeToken)
	req.Header.Set("albumuuid", albmnId)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ChibisafeResponse{}, fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return ChibisafeResponse{}, fmt.Errorf("upload failed: %s", respBody)
	}

	var chibiResp ChibisafeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chibiResp); err != nil {
		return ChibisafeResponse{}, fmt.Errorf("failed to parse chibisafe response: %w", err)
	}

	if len(chibiResp.UUID) == 0 {
		return ChibisafeResponse{}, fmt.Errorf("chibisafe response does not contain UUID")
	}

	return ChibisafeResponse{
		Name:       chibiResp.Name,
		UUID:       chibiResp.UUID,
		URL:        chibiResp.URL,
		Identifier: chibiResp.Identifier,
		PublicURL:  chibiResp.PublicURL,
	}, nil
}

func (s *MediaService) DeleteFromChibisafe(image *models.ImageUpload) error {
	cfg := config.LoadConfig()
	chibisafeURL := cfg.ChibisafeURL
	chibisafeToken := cfg.ChibisafeKey

	// Create request to delete image
	req, err := http.NewRequest("DELETE", chibisafeURL+"/api/admin/file/"+image.FileID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("x-api-key", chibisafeToken)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s", respBody)
	}

	err = s.Repo.DeleteByID(image.ID.String())
	if err != nil {
		return fmt.Errorf("failed to delete from database: %w", err)
	}

	return nil
}

func (s *MediaService) GetImagesByPostID(postID uuid.UUID) ([]models.ImageUpload, error) {
	return s.Repo.GetImagesByPostID(postID)
}

func (s *MediaService) UpdateImageUsage(image *models.ImageUpload) error {
	return s.Repo.UpdateImageUsage(image)
}

func (s *MediaService) DeleteUnusedImages() error {
	// Step 1: ดึงเฉพาะรูปที่ is_used = false และ file_id ไม่ซ้ำใน DB
	unusedImages, err := s.Repo.FindUnusedWithUniqueFileID()
	if err != nil {
		return fmt.Errorf("failed to find unused images: %w", err)
	}

	for _, img := range unusedImages {
		s.Logger.Info("Deleting unused image", zap.String("image_id", img.ID.String()), zap.String("file_id", img.FileID))

		// ลบจาก Chibisafe
		err := s.DeleteFromChibisafe(&img)
		if err != nil {
			s.Logger.Error("Failed to delete image from Chibisafe", zap.Error(err), zap.String("image_id", img.ID.String()))
			continue // ข้ามหากลบไม่ได้
		}
	}

	// Step 2: ลบจากฐานข้อมูล
	if err := s.Repo.DeleteImagesWhereUnused(); err != nil {
		return fmt.Errorf("failed to delete unused images from database: %w", err)
	}

	s.Logger.Info("Successfully deleted unused images from Chibisafe and database")

	return nil
}

func (s *MediaService) GetImageByURL(imageURL string) (*models.ImageUpload, error) {
	image, err := s.Repo.GetImageByURL(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get image by URL: %w", err)
	}
	return image, nil
}
