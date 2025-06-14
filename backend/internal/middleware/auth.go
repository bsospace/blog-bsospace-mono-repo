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

// AuthMiddleware ใช้สำหรับป้องกัน API ด้วย JWT
func AuthMiddleware(userService *user.Service, cryptoService *crypto.CryptoService, cache *cache.Service, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			logger.Error("[ERROR] No token provided in request")
			response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "No token provided")
			c.Abort()
			return
		}

		// ตรวจสอบ Token ผ่าน CryptoService (ใช้ Dependency Injection)
		claims, err := verifyToken(tokenString, cryptoService, logger)

		if err != nil {
			logger.Error("[ERROR] Invalid token", zap.Error(err))
			response.JSONError(c, http.StatusUnauthorized, "Unauthorized", err.Error())
			c.Abort()
			return
		}

		// ค้นหา User ใน Cache ก่อน
		userCache, err := cache.GetUserCache(claims.Email)

		if err == nil && userCache != nil {
			// ถ้า User อยู่ใน Cache ให้เพิ่มข้อมูล User ลง Context
			c.Set("user", userCache)
			logger.Info("[INFO] User found in cache:", zap.String("email", claims.Email))
			c.Next()
			return
		} else {
			logger.Info("[INFO] User not found in cache, checking database:", zap.String("email", claims.Email))
		}

		// ค้นหา User ใน Database
		userDB, err := userService.GetUserByEmail(claims.Email)

		if userDB != nil {
			// set user in cache
			if err := cache.SetUserCache(userDB.Email, userDB); err != nil {
				logger.Error("[ERROR] Failed to set user in cache", zap.Error(err), zap.String("email", userDB.Email))
			} else {
				logger.Info("[INFO] User cached successfully", zap.String("email", userDB.Email))
			}
		}

		if err != nil {

			// Get User Data from OpenId
			userOpenID, err := userService.GetUserProfileOpenId(tokenString)

			if err != nil {
				response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "Failed to get user profile from OpenID")
				logger.Error("[ERROR] Failed to get user profile from OpenID", zap.Error(err))
				c.Abort()
				return
			}

			// ถ้าไม่พบ User ใน Database ให้สร้าง User ใหม่
			newUser := &models.User{
				Email:    claims.Email,
				Avatar:   userOpenID.Image,
				UserName: strings.Split(claims.Email, "@")[0],
				Role:     models.UserRole("NORMAL_USER"),
			}

			if _, err = userService.RegisterUser(newUser); err != nil {
				logger.Error("[ERROR] Failed to register new user", zap.Error(err), zap.String("email", newUser.Email))
				response.JSONError(c, http.StatusInternalServerError, "Internal Server Error", "Failed to register new user")
				c.Abort()
				return
			}

			logger.Info("[INFO] New user registered successfully", zap.String("email", newUser.Email))
			userDB = newUser

			// เพิ่ม User ใหม่ลง Cache
			if err := cache.SetUserCache(userDB.Email, userDB); err != nil {
				logger.Error("[ERROR] Failed to cache new user", zap.Error(err), zap.String("email", userDB.Email))
			} else {
				logger.Info("[INFO] New user cached successfully", zap.String("email", userDB.Email))
			}
		}

		// เพิ่มข้อมูล User ลง Context
		c.Set("user", userDB)
		c.Next()
	}
}

// Extract Token จาก Header หรือ Cookie
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")

	// ตรวจสอบว่า Header มีค่าเป็น Bearer Token หรือไม่
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	if cookie, err := c.Cookie("accessToken"); err == nil {
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
