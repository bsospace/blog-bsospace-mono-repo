package post

import (
	"errors"
	"net/http"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/pkg/errs"
	"rag-searchbot-backend/pkg/ginctx"
	"rag-searchbot-backend/pkg/response"

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
		response.JSONError(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
	}

	if err := h.service.CreatePost(post, user); err != nil {
		response.JSONError(c, http.StatusInternalServerError, "Failed to create post", err.Error())
		return
	}

	response.JSONSuccess(c, 201, "Post created successfully", nil)
}

func (h *PostHandler) GetByShortSlug(c *gin.Context) {

	shortSlug := c.Param("short_slug")

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
	}

	post, err := h.service.GetByShortSlug(shortSlug + "-" + user.ID.String())
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

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
	}

	posts, err := h.service.MyPosts(user)
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

func (h *PostHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
	}

	err := h.service.DeletePostByID(id, user)

	if errors.Is(err, errs.ErrPostNotFound) {
		response.JSONError(c, http.StatusNotFound, "Post not found", "The post you are trying to delete does not exist")
		return
	}

	if errors.Is(err, errs.ErrUnauthorized) {
		response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "You are not the author of this post")
		return
	}

	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "Failed to delete post", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PostHandler) Publish(c *gin.Context) {

	var post post.PublishPostRequestDTO
	var shortSlug = c.Param("short_slug")

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

	published := h.service.PublishPost(&post, userData, shortSlug)

	if published != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to publish post",
			"error":   published.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Post published successfully",
		"data":    post,
	})
}

func (h *PostHandler) Unpublish(c *gin.Context) {

	var shortSlug = c.Param("short_slug")

	if shortSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Short slug is required",
			"error":   "Short slug cannot be empty",
		})
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

	unpublished := h.service.UnpublishPost(userData, shortSlug)

	if unpublished != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to unpublish post",
			"error":   unpublished.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Post unpublished successfully",
	})
}
func (h *PostHandler) GetPublicPostBySlugAndUsername(c *gin.Context) {
	slug := c.Param("slug")
	username := c.Param("username")

	post, err := h.service.GetPublicPostBySlugAndUsername(slug, username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "post not found",
			"message": err.Error(),
		})
		return
	}

	response := MapGetPublicPostBySlugAndUsernameResponse(post)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Get public post by slug and username successfully.",
	})
}
