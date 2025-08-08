package chat

import (
	"rag-searchbot-backend/internal/models"
	"time"

	"github.com/google/uuid"
)

type ServiceInterface interface {
	SendMessage(msg *models.AIResponse) error
	GetChatBetweenUsers(user1, user2 uuid.UUID, limit, offset int) ([]models.AIResponse, error)
	GetChatByRoom(roomID uuid.UUID, limit, offset int) ([]models.AIResponse, error)
	MarkMessageAsRead(messageID uuid.UUID, readAt time.Time) error
}

type Service struct {
	Repo RepositoryInterface
}

func NewService(repo RepositoryInterface) ServiceInterface {
	return &Service{Repo: repo}
}

func (s *Service) SendMessage(msg *models.AIResponse) error {
	return s.Repo.CreateMessage(msg)
}

func (s *Service) GetChatBetweenUsers(user1, user2 uuid.UUID, limit, offset int) ([]models.AIResponse, error) {
	return s.Repo.GetMessagesBetweenUsers(user1, user2, limit, offset)
}

func (s *Service) GetChatByRoom(roomID uuid.UUID, limit, offset int) ([]models.AIResponse, error) {
	return s.Repo.GetMessagesByRoom(roomID, limit, offset)
}

func (s *Service) MarkMessageAsRead(messageID uuid.UUID, readAt time.Time) error {
	return s.Repo.MarkAsRead(messageID, readAt)
}
