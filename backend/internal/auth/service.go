package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"
)

type AuthServiceInterface interface {
	ExchangeToken(code string) (*TokenExchangeResponse, error)
}

type AuthService struct {
	UserService    user.ServiceInterface
	CrypetoService *crypto.CryptoService
	EnvConfig      *config.Config
}

func NewAuthService(userService user.ServiceInterface, cryptoService *crypto.CryptoService, env *config.Config) AuthServiceInterface {
	return &AuthService{
		UserService:    userService,
		CrypetoService: cryptoService,
		EnvConfig:      env,
	}
}

// TokenExchangeResponse ควรสอดคล้องกับ JSON ที่ backend ส่งกลับมา
type TokenExchangeResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"openid.atk"`
	RefreshToken string `json:"openid.rtk"`
}

// Exchange token Open ID Oauth2.0
func (s *AuthService) ExchangeToken(code string) (*TokenExchangeResponse, error) {
	if code == "" {
		return nil, errors.New("code is required")
	}

	req, err := http.NewRequest("GET", s.EnvConfig.OpenIDURL+"/auth/exchange?token="+code, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to exchange token, status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !tokenResp.Success || tokenResp.AccessToken == "" {
		return nil, errors.New("missing or invalid token in response")
	}

	return &tokenResp, nil
}
