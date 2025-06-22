package notification

import (
	"net/http"
	"rag-searchbot-backend/internal/notification"
	"rag-searchbot-backend/pkg/ginctx"
	"rag-searchbot-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	NotiService notification.NotificationService
}

func NewNotificationHandler(notiService *notification.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		NotiService: *notiService,
	}
}

// GetNotificationsHandler retrieves notifications for the authenticated user

func (h *NotificationHandler) GetNotificationsHandler(c *gin.Context) {

	// pagination
	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")

	// validate limit and page
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		response.JSONError(c, http.StatusBadRequest, "Invalid limit", "Limit must be a positive integer")
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		response.JSONError(c, http.StatusBadRequest, "Invalid page", "Page must be a positive integer")
		return
	}

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
	}

	notifications, err := h.NotiService.GetNotifications(*user, limitStr, pageStr)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "Failed to retrieve notifications", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    notifications,
	})
}

// MarkNotificationAsReadHandler marks a notification as read
func (h *NotificationHandler) MarkNotificationAsReadHandler(c *gin.Context) {
	notiIDStr := c.Param("id")
	notiID, err := strconv.Atoi(notiIDStr)
	if err != nil || notiID <= 0 {
		response.JSONError(c, http.StatusBadRequest, "Invalid notification ID", "Notification ID must be a positive integer")
		return
	}

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
		return
	}

	err = h.NotiService.MarkAsRead(uint(notiID), *user)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "Failed to mark notification as read", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification marked as read",
	})
}

// MarkAllNotificationsAsReadHandler marks all notifications as read
func (h *NotificationHandler) MarkAllNotificationsAsReadHandler(c *gin.Context) {
	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
		return
	}

	err := h.NotiService.MarkAllAsRead(*user)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "Failed to mark all notifications as read", err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusOK, "All notifications marked as read", nil)
}
