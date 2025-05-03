package middleware

import (
	"errors"
	"fmt"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware ใช้สำหรับป้องกัน API ด้วย JWT
func AuthMiddleware(userService *user.Service, cryptoService *crypto.CryptoService, cache *cache.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "Missing token",
			})
			c.Abort()
			return
		}

		// ตรวจสอบ Token ผ่าน CryptoService (ใช้ Dependency Injection)
		claims, err := verifyToken(tokenString, cryptoService)

		fmt.Printf("claims:%+v\n", claims)
		fmt.Print("err: \n", err)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// ค้นหา User ใน Database
		userDB, err := userService.GetUserByEmail(claims.Email)

		fmt.Printf("userDB: %+v\n", userDB)
		fmt.Printf("err: %v\n", err)

		if err != nil {

			// Get User Data from OpenId
			userOpenID, err := userService.GetUserProfileOpenId(tokenString)

			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user profile", "message": err.Error()})
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
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to register user"})
				c.Abort()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// เพิ่มข้อมูล User ลง Context
		c.Set("user", userDB)
		c.Next()
	}
}

// Extract Token จาก Header หรือ Cookie
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")

	fmt.Println("authHeader: ", authHeader)
	// ตรวจสอบว่า Header มีค่าเป็น Bearer Token หรือไม่
	// fmt.Println("authHeader: ", authHeader)
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
func verifyToken(tokenString string, cryptoService *crypto.CryptoService) (*TokenClaims, error) {

	fmt.Println("tokenString: ", tokenString)
	token, err := cryptoService.SmartVerifyToken(tokenString, "Access")
	if err != nil {
		log.Println("Invalid token:", err)
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
