package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/models"

	"github.com/google/uuid"
)

type MediaService struct {
	Repo *MediaRepository
}

func NewMediaService(repo *MediaRepository) *MediaService {
	return &MediaService{Repo: repo}
}

func (s *MediaService) CreateMedia(fileHeader *multipart.FileHeader, user *models.User, postID *uuid.UUID) (*models.ImageUpload, error) {

	// Upload ไปยัง Chibisafe (สมมุติ)
	res, err := s.UploadToChibisafe(fileHeader)
	if err != nil {
		return nil, err
	}

	log.Println("DEBUG - user.ID:", user.ID)
	log.Println("RESULT - Chibisafe Response:", res)

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

func (s *MediaService) UploadToChibisafe(fileHeader *multipart.FileHeader) (ChibisafeResponse, error) {

	// log config

	cfg := config.LoadConfig()
	chibisafeURL := cfg.ChibisafeURL
	chibisafeToken := cfg.ChibisafeKey
	albmnId := cfg.ChibisafeAlbumId

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		return ChibisafeResponse{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Prepare multipart/form-data body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileHeader.Filename+"-"+uuid.New().String())
	if err != nil {
		return ChibisafeResponse{}, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
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

	// Parse response

	var chibiResp ChibisafeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chibiResp); err != nil {
		return ChibisafeResponse{}, fmt.Errorf("failed to parse chibisafe response: %w", err)
	}

	if len(chibiResp.UUID) == 0 {
		return ChibisafeResponse{}, fmt.Errorf("chibisafe response does not contain UUID")
	}

	// Return full URL
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
		log.Println("Deleting from chibisafe:", img.ImageURL)

		// ลบจาก Chibisafe
		err := s.DeleteFromChibisafe(&img)
		if err != nil {
			log.Printf("warning: failed to delete image %s from chibisafe: %v", img.ID, err)
			continue // ข้ามหากลบไม่ได้
		}
	}

	// Step 2: ลบจากฐานข้อมูล
	if err := s.Repo.DeleteImagesWhereUnused(); err != nil {
		return fmt.Errorf("failed to delete unused images from database: %w", err)
	}

	log.Println("Successfully deleted unused images from Chibisafe and database")

	return nil
}
