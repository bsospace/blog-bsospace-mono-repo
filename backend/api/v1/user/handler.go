package user

import (
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/ginctx"
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

type UpdateUserRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}

// Update user data
func (h *UserHandler) UpdateUser(c *gin.Context) {
	user, ok := ginctx.GetUserFromContext(c)

	if !ok || user == nil {
		response.JSONError(c, 401, "Unauthorized", "User not found in context")
		return
	}

	// ดึงข้อมูลที่ต้องการอัพเดตจาก request body
	var updateData UpdateUserRequest

	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.JSONError(c, 400, "Bad Request", "Invalid request data")
		return
	}

	// สร้าง user object ใหม่จากข้อมูลที่ได้รับ
	user.UserName = updateData.Username
	user.FirstName = updateData.FirstName
	user.LastName = updateData.LastName
	user.Bio = updateData.Bio
	updatedUser, err := h.userService.UpdateUser(user)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.JSONError(c, 404, "User not found", "User does not exist")
			return
		}
		response.JSONError(c, 500, "Internal Server Error", err.Error())
		return
	}
	response.JSONSuccess(c, 200, "User updated successfully", updatedUser)
}
