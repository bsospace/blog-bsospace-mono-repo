package ai

import (
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"
)

type AIService struct {
	PosRepo      post.PostRepositoryInterface
	TaskEnqueuer *TaskEnqueuer
}

func NewAIService(posRepo post.PostRepositoryInterface, enqueuer *TaskEnqueuer) *AIService {
	return &AIService{
		PosRepo:      posRepo,
		TaskEnqueuer: enqueuer,
	}
}

func (s *AIService) OpenAIMode(postID string, userData *models.User) (bool, error) {
	post, err := s.PosRepo.GetByID(postID)
	if err != nil {
		return false, err
	}
	if post == nil {
		return false, nil
	}

	if post.AIChatOpen {
		return false, nil // AI chat already open
	}

	if post.AuthorID != userData.ID {
		return false, nil // User is not the author of the post
	}

	_, err = s.TaskEnqueuer.EnqueuePostEmbedding(post, userData)
	if err != nil {
		return false, err
	}

	return true, nil
}

// DisableOpenAIMode disables the OpenAI mode for a post
func (s *AIService) DisableOpenAIMode(postID string, userData *models.User) (bool, error) {
	post, err := s.PosRepo.GetByID(postID)
	if err != nil {
		return false, err
	}
	if post == nil {
		return false, nil
	}

	if !post.AIChatOpen {
		return false, nil // AI chat already disabled
	}

	if post.AuthorID != userData.ID {
		return false, nil // User is not the author of the post
	}

	post.AIChatOpen = false
	err = s.PosRepo.Update(post)
	if err != nil {
		return false, err
	}

	return true, nil
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"` // false = return full output
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}
