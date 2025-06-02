package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"rag-searchbot-backend/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// CryptoService manages cryptographic functions
type CryptoService struct {
	KeyDirectory string
}

// NewCryptoService initializes a new CryptoService
func NewCryptoService() *CryptoService {
	return &CryptoService{
		KeyDirectory: filepath.Join("../../keys"),
	}
}

// HashPassword hashes a plain-text password
func (cs *CryptoService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ComparePasswords compares a plain-text password with a hashed password
func (cs *CryptoService) ComparePasswords(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// getKeyPath constructs the file path for a key
func (cs *CryptoService) getKeyPath(service, keyType, keyScope string) string {
	absPath, err := filepath.Abs(cs.KeyDirectory)
	if err != nil {
		// If we can't get absolute path, fall back to relative path
		return filepath.Join(cs.KeyDirectory, fmt.Sprintf("%s%s%s.pem", service, keyScope, keyType))
	}
	return filepath.Join(absPath, fmt.Sprintf("%s%s%s.pem", service, keyScope, keyType))
}

// readKey reads the key file content
func (cs *CryptoService) readKey(service, keyType, keyScope string) ([]byte, error) {
	keyPath := cs.getKeyPath(service, keyType, keyScope)
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s %s key for service '%s': %w", keyType, keyScope, service, err)
	}
	return key, nil
}

// generateToken generates a JWT token
func (cs *CryptoService) generateToken(userID, username, service, keyType string, expiry time.Duration) (string, error) {
	privateKeyData, err := cs.readKey(service, keyType, "Private")
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return "", errors.New("failed to decode private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return "", err
	}

	cfg := config.LoadConfig()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":     userID,
		"name":    username,
		"iss":     cfg.AppUrl,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(expiry).Unix(),
		"service": service,
	})

	return token.SignedString(privateKey)
}

// GenerateAccessToken creates an access token for a user
func (cs *CryptoService) GenerateAccessToken(userID, username, service string) (string, error) {
	return cs.generateToken(userID, username, service, "Access", 24*time.Hour)
}

// GenerateRefreshToken creates a refresh token for a user
func (cs *CryptoService) GenerateRefreshToken(userID, username, service string) (string, error) {
	return cs.generateToken(userID, username, service, "Refresh", 15*24*time.Hour)
}

// VerifyToken verifies a JWT token
func (cs *CryptoService) SmartVerifyToken(tokenString, keyType string) (*jwt.Token, error) {
	// แกะ token ก่อน verify เพื่อดูว่าเป็นของ service ไหน
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token for service detection: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims format")
	}

	service, ok := claims["service"].(string)
	if !ok || service == "" {
		return nil, errors.New("missing or invalid service in token")
	}

	// อ่าน public key จาก service ที่เจอใน claims
	publicKeyData, err := cs.readKey(service, keyType, "Public")
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(publicKeyData)
	if block == nil {
		return nil, errors.New("failed to decode public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// parse แบบ verify จริง
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})
}

// DecodeToken decodes a JWT token without verifying
func (cs *CryptoService) DecodeToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, nil)
}
