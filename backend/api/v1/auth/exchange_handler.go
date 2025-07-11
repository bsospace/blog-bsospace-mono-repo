package auth

import (
	"net/http"
	"rag-searchbot-backend/internal/auth"
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/models"
	"strings"

	"github.com/gin-gonic/gin"
)

func Exchange(container *container.Container) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			var req struct {
				Code string `json:"code" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Missing or invalid code"})
				return
			}
			code = req.Code
		}

		authService := auth.NewAuthService(container.UserService, container.CryptoService, container.Env)
		result, err := authService.(*auth.AuthService).ExchangeToken(code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error()})
			return
		}

		// Set cookies
		isProd := container.Env.AppEnv == "release"
		domains := strings.Split(container.Env.Domain, ",")

		for _, domain := range domains {
			domain = strings.TrimSpace(domain)
			c.SetCookie("blog.atk", result.AccessToken, 3600, "/", domain, isProd, true)
			c.SetCookie("blog.rtk", result.RefreshToken, 3600*24*7, "/", domain, isProd, true)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": result.Message,
			"data": gin.H{
				"access_token":  result.AccessToken,
				"refresh_token": result.RefreshToken,
			},
		})
	}
}

// log out
func Logout(container *container.Container) gin.HandlerFunc {
	return func(c *gin.Context) {
		isProd := container.Env.AppEnv == "release"
		domains := strings.Split(container.Env.Domain, ",")

		// Clear cookies
		for _, domain := range domains {
			domain = strings.TrimSpace(domain)
			c.SetCookie("blog.atk", "", -1, "/", domain, isProd, true)
			c.SetCookie("blog.rtk", "", -1, "/", domain, isProd, true)
		}

		// Get user from context
		userVal, exists := c.Get("user")
		if exists && userVal != nil {
			if user, ok := userVal.(*models.User); ok && user != nil {
				// Clear user cache
				container.CacheService.ClearUserCache(user.Email)

				// Clear warp key
				container.CacheService.ClearWarpKey(user.Email)
			}
		}

		// Respond
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Logged out successfully",
		})
	}
}
