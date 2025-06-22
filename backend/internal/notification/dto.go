package notification

import (
	"rag-searchbot-backend/internal/models"
	"time"
)

type GetNotificationDTO struct {
	ID      uint      `json:"id"`
	Title   string    `json:"title"`
	Even    string    `json:"event"`
	Content string    `json:"content"`
	Link    string    `json:"link"`
	Seen    bool      `json:"seen"`
	SeenAt  time.Time `json:"seen_at,omitempty"`
}

func MapNotificationToDTO(noti models.Notification) GetNotificationDTO {
	return GetNotificationDTO{
		ID:      noti.ID,
		Title:   noti.Title,
		Even:    noti.Even,
		Content: noti.Content,
		Link:    noti.Link,
		Seen:    noti.Seen,
		SeenAt:  noti.SeenAt,
	}
}

type NotificationRepositoryQuery struct {
	Limit        int                   `json:"limit"`
	Total        int64                 `json:"total"`
	HasNext      bool                  `json:"has_next"`
	Page         int                   `json:"page"`
	Offset       int                   `json:"offset"`
	Search       string                `json:"search"`
	Notification []models.Notification `json:"notification"`
}

type NotiListResponse struct {
	Notification []GetNotificationDTO `json:"notification"`
	Meta         Meta                 `json:"meta"`
}

type Meta struct {
	Total       int64 `json:"total"`
	HasNextPage bool  `json:"hasNextPage"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPage   int   `json:"totalPage"`
}
