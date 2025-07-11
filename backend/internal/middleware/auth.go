package middleware

import (
	"errors"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"
	"rag-searchbot-backend/pkg/response"

	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthMiddleware struct for DI
//go:generate mockgen -destination=../../mocks/mock_authmiddleware.go -package=mocks rag-searchbot-backend/internal/middleware AuthMiddleware

type AuthMiddleware struct {
	UserService   user.ServiceInterface
	CryptoService *crypto.CryptoService
	CacheService  cache.ServiceInterface
	Logger        *zap.Logger
}

func NewAuthMiddleware(userService user.ServiceInterface, cryptoService *crypto.CryptoService, cacheService cache.ServiceInterface, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		UserService:   userService,
		CryptoService: cryptoService,
		CacheService:  cacheService,
		Logger:        logger,
	}
}

func (a *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			a.Logger.Error("[ERROR] No token provided in request")
			response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "No token provided")
			c.Abort()
			return
		}

		claims, err := verifyToken(tokenString, a.CryptoService, a.Logger)
		if err != nil {
			a.Logger.Error("[ERROR] Invalid token", zap.Error(err))
			response.JSONError(c, http.StatusUnauthorized, "Unauthorized", err.Error())
			c.Abort()
			return
		}

		var userDB *models.User

		// Try cache first
		userCache, err := a.CacheService.GetUserCache(claims.Email)
		if err == nil && userCache != nil {
			userDB = userCache
			a.Logger.Info("[INFO] User found in cache", zap.String("email", claims.Email))
		} else {
			a.Logger.Info("[INFO] User not found in cache, checking database", zap.String("email", claims.Email))

			userDB, err = a.UserService.GetUserByEmail(claims.Email)
			if userDB != nil {
				// Cache it
				if err := a.CacheService.SetUserCache(userDB.Email, userDB); err != nil {
					a.Logger.Error("[ERROR] Failed to set user in cache", zap.Error(err), zap.String("email", userDB.Email))
				} else {
					a.Logger.Info("[INFO] User cached successfully", zap.String("email", userDB.Email))
				}
			}

			// Register new user if not exist
			if err != nil {
				userOpenID, err := a.UserService.GetUserProfileOpenId(tokenString)
				if err != nil {
					response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "Failed to get user profile from OpenID")
					a.Logger.Error("[ERROR] Failed to get user profile from OpenID", zap.Error(err))
					c.Abort()
					return
				}

				newUser := &models.User{
					Email:    claims.Email,
					Avatar:   userOpenID.Image,
					UserName: strings.Split(claims.Email, "@")[0],
					Role:     models.UserRole("NORMAL_USER"),
				}

				if _, err = a.UserService.RegisterUser(newUser); err != nil {
					a.Logger.Error("[ERROR] Failed to register new user", zap.Error(err), zap.String("email", newUser.Email))
					response.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "Failed to register new user")
					c.Abort()
					return
				}

				a.Logger.Info("[INFO] New user registered successfully", zap.String("email", newUser.Email))
				userDB = newUser

				if err := a.CacheService.SetUserCache(userDB.Email, userDB); err != nil {
					a.Logger.Error("[ERROR] Failed to cache new user", zap.Error(err), zap.String("email", userDB.Email))
				} else {
					a.Logger.Info("[INFO] New user cached successfully", zap.String("email", userDB.Email))
				}
			}
		}

		// Generate OTP warp key
		otpToken, err := a.CryptoService.GenerateSocketToken()
		if err != nil {
			a.Logger.Error("[ERROR] Failed to generate OTP token", zap.Error(err))
			response.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "Failed to generate OTP token")
			c.Abort()
			return
		}

		// Clear old warp key and set new one
		a.CacheService.ClearWarpKey(userDB.Email)
		if err := a.CacheService.SetWarpKey(userDB.Email, otpToken); err != nil {
			a.Logger.Error("[ERROR] Failed to cache warp key", zap.Error(err), zap.String("email", userDB.Email))
		}

		// Set to Gin context
		c.Set("user", userDB)
		c.Set("warp_key", otpToken)

		c.Next()
	}
}

func SocketAuthMiddleware(userService *user.Service, cryptoService *crypto.CryptoService, cache *cache.Service, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("warp")

		if token == "" {
			logger.Error("[ERROR] No token provided in request")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			return
		}

		// Get email from warp key
		email, exists := cache.GetWarpEmail(token)
		if !exists || email == "" {
			logger.Error("[ERROR] Invalid warp key", zap.String("warp_key", token))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid warp key"})
			return
		}

		// Load user from cache or DB
		userCache, err := cache.GetUserCache(email)
		if err == nil && userCache != nil {
			c.Set("user", userCache)
			c.Next()
			return
		}

		userDB, err := userService.GetUserByEmail(email)
		if err != nil || userDB == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Set to cache for next time
		if err := cache.SetUserCache(userDB.Email, userDB); err != nil {
			logger.Error("[ERROR] Failed to cache user", zap.Error(err))
		}

		c.Set("user", userDB)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// allow cookie only
	if cookie, err := c.Cookie("blog.atk"); err == nil {
		return cookie
	}
	return ""
}

// Token Claims
type TokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// Verify JWT Token (ใช้ Dependency Injection)
func verifyToken(tokenString string, cryptoService *crypto.CryptoService, logger *zap.Logger) (*TokenClaims, error) {

	token, err := cryptoService.SmartVerifyToken(tokenString, "Access")
	if err != nil {
		logger.Error("[ERROR] Failed to verify token", zap.Error(err))
		return nil, err
	}

	// log.Println("Valid token:", token)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &TokenClaims{
		Email: claims["email"].(string),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    claims["iss"].(string),
			Subject:   claims["sub"].(string),
			ExpiresAt: jwt.NewNumericDate(time.Unix(int64(claims["exp"].(float64)), 0)),
			IssuedAt:  jwt.NewNumericDate(time.Unix(int64(claims["iat"].(float64)), 0)),
		},
	}, nil
}
