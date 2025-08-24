package middleware

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OptionalAuthMiddleware struct for DI
type OptionalAuthMiddleware struct {
	UserService   user.ServiceInterface
	CryptoService *crypto.CryptoService
	CacheService  cache.ServiceInterface
	Logger        *zap.Logger
}

func NewOptionalAuthMiddleware(userService user.ServiceInterface, cryptoService *crypto.CryptoService, cacheService cache.ServiceInterface, logger *zap.Logger) *OptionalAuthMiddleware {
	return &OptionalAuthMiddleware{
		UserService:   userService,
		CryptoService: cryptoService,
		CacheService:  cacheService,
		Logger:        logger,
	}
}

// Handler is an optional authentication middleware
// It doesn't require authentication but stores user context if valid token is provided
func (o *OptionalAuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c, "blog.atk")
		if tokenString == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		claims, err := verifyToken(tokenString, o.CryptoService, o.Logger)
		if err != nil {
			// Invalid token, continue without authentication
			o.Logger.Warn("[WARN] Invalid token in optional auth", zap.Error(err))
			c.Next()
			return
		}

		var userDB *models.User

		// Try cache first
		userCache, err := o.CacheService.GetUserCache(claims.Email)
		if err == nil && userCache != nil {
			userDB = userCache
			o.Logger.Info("[INFO] User found in cache (optional auth)", zap.String("email", claims.Email))
		} else {
			o.Logger.Info("[INFO] User not found in cache, checking database (optional auth)", zap.String("email", claims.Email))

			userDB, _ = o.UserService.GetUserByEmail(claims.Email)
			if userDB != nil {
				// Cache it
				if err := o.CacheService.SetUserCache(userDB.Email, userDB); err != nil {
					o.Logger.Error("[ERROR] Failed to set user in cache (optional auth)", zap.Error(err), zap.String("email", userDB.Email))
				} else {
					o.Logger.Info("[INFO] User cached successfully (optional auth)", zap.String("email", userDB.Email))
				}
			} else {
				// User not found, continue without authentication
				o.Logger.Warn("[WARN] User not found in database (optional auth)", zap.String("email", claims.Email))
				c.Next()
				return
			}
		}

		// Set user to Gin context
		c.Set("user", userDB)
		c.Set("user_id", userDB.ID)

		c.Next()
	}
}
