package post

import (
	"net/http"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	service *post.PostService
}

func NewPostHandler(service *post.PostService) *PostHandler {
	return &PostHandler{service: service}
}

func (h *PostHandler) Create(c *gin.Context) {
	var post post.CreatePostRequest
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found in context",
		})
		return
	}

	userData, ok := user.(*models.User)
	if !ok || userData == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user data",
		})
		return
	}

	if err := h.service.CreatePost(post, userData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create post",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Post created successfully",
	})
}

func (h *PostHandler) GetByShortSlug(c *gin.Context) {

	shortSlug := c.Param("short_slug")

	user, exists := c.Get("user")
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found in context",
		})
		return
	}

	userData, ok := user.(*models.User)
	if !ok || userData == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user data",
		})
		return
	}

	post, err := h.service.GetByShortSlug(shortSlug + "-" + userData.ID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    post,
		"message": "Get post by short slug successfully.",
	})
}

func (h *PostHandler) Update(c *gin.Context) {

}

func (h *PostHandler) GetAll(c *gin.Context) {
	posts, err := h.service.GetPosts(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    posts,
		"message": "Get all posts successfully.",
	})
}

func (h *PostHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	post, err := h.service.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	post, err := h.service.GetPostBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "post not found",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    post,
		"message": "Get post by slug successfully.",
	})
}

func (h *PostHandler) MyPost(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found in context",
		})
		return
	}

	userData, ok := user.(*models.User)
	if !ok || userData == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user data",
		})
		return
	}

	posts, err := h.service.MyPosts(userData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch posts",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Fetched posts successfully",
		"data":    posts,
	})
}

// func (h *PostHandler) Update(c *gin.Context) {
// 	id := c.Param("id")
// 	var post models.Post
// 	if err := c.ShouldBindJSON(&post); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	post.ID = post.ID // keep existing UUID
// 	if err := h.service.UpdatePost(&post); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, post)
// }

func (h *PostHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
