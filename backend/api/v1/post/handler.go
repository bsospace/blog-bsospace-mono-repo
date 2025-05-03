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
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.CreatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, post)
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
