package user

import (
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService *user.Service
}

func NewUserHandler(userService *user.Service) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetExistingUsername(c *gin.Context) {

	// ดึง username จาก query parameter
	username := c.Query("username")
	if username == "" {
		response.JSONError(c, 400, "Bad Request", "Username is required")
		return
	}

	// ใช้ userService เพื่อตรวจสอบว่า username มีอยู่ในระบบหรือไม่
	user, err := h.userService.GetExistingUsername(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.JSONError(c, 404, "Username not found", "Username does not exist")
			return
		}
		response.JSONError(c, 500, "Internal Server Error", err.Error())
		return
	}
	response.JSONSuccess(c, 200, "Success", user)
}
