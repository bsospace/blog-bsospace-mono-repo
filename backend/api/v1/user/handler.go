package user

import (
	"strings"

	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/ginctx"
	"rag-searchbot-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService *user.Service
	postService post.PostServiceInterface
}

func NewUserHandler(userService *user.Service, postService post.PostServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
		postService: postService,
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
	Avatar    string `json:"avatar,omitempty"`
	Location  string `json:"location,omitempty"`
	Website   string `json:"website,omitempty"`
	GitHub    string `json:"github,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	Facebook  string `json:"facebook,omitempty"`
	YouTube   string `json:"youtube,omitempty"`
	Discord   string `json:"discord,omitempty"`
	Telegram  string `json:"telegram,omitempty"`
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
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	// สร้าง user object ใหม่จากข้อมูลที่ได้รับ
	user.UserName = updateData.Username
	user.FirstName = updateData.FirstName
	user.LastName = updateData.LastName
	user.Bio = updateData.Bio
	user.Avatar = updateData.Avatar
	user.Location = updateData.Location
	user.Website = updateData.Website
	user.GitHub = updateData.GitHub
	user.Twitter = updateData.Twitter
	user.LinkedIn = updateData.LinkedIn
	user.Instagram = updateData.Instagram
	user.Facebook = updateData.Facebook
	user.YouTube = updateData.YouTube
	user.Discord = updateData.Discord
	user.Telegram = updateData.Telegram
	user.NewUser = false

	err := h.userService.UpdateUser(user)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.JSONError(c, 404, "User not found", "User does not exist")
			return
		}
		response.JSONError(c, 500, "Internal Server Error", err.Error())
		return
	}
	response.JSONSuccess(c, 200, "User updated successfully", nil)
}

// GetUserProfile ดึงข้อมูล User Profile และบทความที่เขียน
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.JSONError(c, 400, "Bad Request", "Username is required")
		return
	}

	// Get current user ID from context if authenticated
	var currentUserID *uuid.UUID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uuid.UUID); ok {
			currentUserID = &id
		}
	}

	userProfile, err := h.userService.GetUserProfileByUsername(username, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "not found") {
			response.JSONError(c, 404, "User not found", "User does not exist")
			return
		}
		response.JSONError(c, 500, "Internal Server Error", err.Error())
		return
	}

	response.JSONSuccess(c, 200, "User profile retrieved successfully", userProfile)
}

// GetUserProfileWithPosts ดึงข้อมูล User Profile และบทความที่เขียน
func (h *UserHandler) GetUserProfileWithPosts(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.JSONError(c, 400, "Bad Request", "Username is required")
		return
	}

	// Get current user ID from context if available (from optional auth middleware)
	var currentUserID *uuid.UUID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uuid.UUID); ok {
			currentUserID = &id
		}
	}

	userProfile, err := h.userService.GetUserProfileByUsername(username, currentUserID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "not found") {
			response.JSONError(c, 404, "User not found", "User does not exist")
			return
		}
		response.JSONError(c, 500, "Internal Server Error", err.Error())
		return
	}

	// Get user posts
	posts, err := h.postService.GetPostsByAuthor(username, 1, 10)
	if err != nil {
		response.JSONError(c, 500, "Internal Server Error", "Failed to get user posts")
		return
	}

	response.JSONSuccess(c, 200, "User profile and posts retrieved successfully", gin.H{
		"user":  userProfile,
		"posts": posts,
	})
}

// GetSupportedRegions ดึงรายการ regions ที่รองรับ
func (h *UserHandler) GetSupportedRegions(c *gin.Context) {
	regions := []string{
		"Thailand", "Japan", "China", "South Korea", "Vietnam",
		"Indonesia", "Malaysia", "Singapore", "Philippines", "India",
		"United States", "United Kingdom", "Germany", "France", "Canada",
		"Australia", "Brazil", "Mexico", "Russia", "South Africa",
	}

	response.JSONSuccess(c, 200, "Supported regions retrieved successfully", regions)
}
