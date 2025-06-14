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

	_, err = s.TaskEnqueuer.EnqueuePostEmbedding(post, userData)
	if err != nil {
		return false, err
	}

	return true, nil
}
