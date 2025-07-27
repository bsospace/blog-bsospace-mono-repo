package ai

import (
	"time"
)

type ChatDTO struct {
	ID        uint      `json:"id"`
	UsedAt    time.Time `json:"used_at"`
	Prompt    string    `json:"prompt"`
	Response  string    `json:"response"`
	TokenUsed int       `json:"token_used"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
