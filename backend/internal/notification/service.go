package notification

import (
	"log"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/ws"
)

type NotificationService struct {
	NotiRepo      *Repository
	SocketManager *ws.Manager
}

func NewService(repo *Repository, socketManager *ws.Manager) *NotificationService {
	return &NotificationService{NotiRepo: repo, SocketManager: socketManager}
}

func (s *NotificationService) Notify(user *models.User, title string, event string, data string, link *string) error {
	noti := &models.Notification{
		Title:   title,
		Content: data,
		UserID:  user.ID,
		Seen:    false,
		Even:    event,
	}

	if link != nil {
		noti.Link = *link
	}

	// Save to DB
	newNoti, err := s.NotiRepo.Create(noti)
	if err != nil {
		return err
	}

	log.Printf("[Notification] New notification for user %s: %s", user.Email, title)

	// Send via WS
	return s.SocketManager.SendToUser(user.ID.String(), event, newNoti)
}
