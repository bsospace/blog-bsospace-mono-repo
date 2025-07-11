package notification

import (
	"log"
	"math"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/ws"
	"rag-searchbot-backend/pkg/errs"
	"strconv"
)

type NotificationServiceInterface interface {
	Notify(user *models.User, title string, event string, data string, link *string) error
}

type NotificationService struct {
	NotiRepo      RepositoryInterface
	SocketManager *ws.Manager
}

func NewService(repo RepositoryInterface, socketManager *ws.Manager) NotificationServiceInterface {
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

// Get notifications for a user
func (s *NotificationService) GetNotifications(user models.User, limitStr, pageStr string) (*NotiListResponse, error) {
	// 1. แปลง limit/page string → int
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// 2. ดึง notifications ตาม limit/offset
	notifications, err := s.NotiRepo.GetByUser(user.ID.String(), limit, offset)
	if err != nil {
		return nil, err
	}

	// 3. นับ total ทั้งหมดของ user นี้
	total, err := s.NotiRepo.CountByUser(user.ID.String())
	if err != nil {
		return nil, err
	}

	// 4. คำนวณ hasNext
	hasNext := int64(offset+limit) < total

	// 5. สร้าง DTO
	result := &NotificationRepositoryQuery{
		Limit:        limit,
		Page:         page,
		Total:        total,
		HasNext:      hasNext,
		Notification: notifications,
	}

	var meta Meta

	meta.Total = result.Total
	meta.HasNextPage = result.HasNext
	meta.Page = result.Page
	meta.Limit = result.Limit
	meta.TotalPage = int(math.Ceil(float64(result.Total) / float64(limit)))

	// Convert []models.Notification to []GetNotificationDTO
	var notificationDTOs []GetNotificationDTO
	for _, noti := range result.Notification {
		dto := GetNotificationDTO{
			ID:      noti.ID,
			Title:   noti.Title,
			Content: noti.Content,
			Seen:    noti.Seen,
			Even:    noti.Even,
			Link:    noti.Link,
		}
		notificationDTOs = append(notificationDTOs, dto)
	}

	return &NotiListResponse{
		Notification: notificationDTOs,
		Meta:         meta,
	}, nil
}

// Mark a notification as read
func (s *NotificationService) MarkAsRead(notiID uint, user models.User) error {
	// Validate notification ID
	noti, err := s.NotiRepo.GetByID(notiID, &user)
	if err != nil {
		return err
	}

	if noti.UserID != user.ID {
		return errs.ErrUnauthorized
	}

	if err := s.NotiRepo.MarkAsRead(notiID); err != nil {
		return err
	}

	log.Printf("[Notification] Notification %d marked as read", notiID)

	return s.SocketManager.SendToUser(noti.UserID.String(), "notification_read", noti)
}

// Mark all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(user models.User) error {
	if err := s.NotiRepo.MarkAllAsRead(user.ID.String()); err != nil {
		return err
	}

	log.Printf("[Notification] All notifications for user %s marked as read", user.Email)

	// Notify via WebSocket
	return s.SocketManager.SendToUser(user.ID.String(), "all_notifications_read", nil)
}

// Delete a notification
func (s *NotificationService) DeleteNotification(notiID uint, user *models.User) error {
	// Validate notification ID
	noti, err := s.NotiRepo.GetByID(notiID, user)
	if err != nil {
		return err
	}

	if noti.UserID != user.ID {
		return errs.ErrUnauthorized
	}

	if err := s.NotiRepo.DeleteByID(notiID, user); err != nil {
		return err
	}

	log.Printf("[Notification] Notification %d deleted for user %s", notiID, user.Email)

	// Notify via WebSocket
	return s.SocketManager.SendToUser(user.ID.String(), "notification_deleted", noti)
}
