package ai

import (
	"encoding/json"
	"rag-searchbot-backend/internal/models"

	"github.com/hibiken/asynq"
)

const TaskTypeEmbedPost = "ai:embed_post"
const TaskTypeFilterPostContentByAI = "ai:filter_post_content"

type EmbedPostPayload struct {
	Post models.Post
	User models.User
}

type TaskEnqueuer struct {
	Client *asynq.Client
}

func NewTaskEnqueuer(client *asynq.Client) *TaskEnqueuer {
	return &TaskEnqueuer{Client: client}
}

func (t *TaskEnqueuer) EnqueuePostEmbedding(post *models.Post, user *models.User) (*asynq.TaskInfo, error) {
	payload, err := json.Marshal(EmbedPostPayload{
		Post: models.Post{
			ID:          post.ID,
			Title:       post.Title,
			HTMLContent: post.HTMLContent,
		},
		User: models.User{
			ID:    user.ID,
			Email: user.Email,
		},
	})
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TaskTypeEmbedPost, payload)
	return t.Client.Enqueue(task)
}

// FilterPostContentByAI
type FilterPostContentByAIPayload struct {
	Post models.Post
	User models.User
}

func (t *TaskEnqueuer) EnqueueFilterPostContentByAI(post *models.Post, user *models.User) (bool, error) {
	payload, err := json.Marshal(FilterPostContentByAIPayload{
		Post: models.Post{
			ID:      post.ID,
			Title:   post.Title,
			Content: post.Content,
		},
		User: models.User{
			ID:    user.ID,
			Email: user.Email,
		},
	})

	if err != nil {
		return false, err
	}
	task := asynq.NewTask(TaskTypeFilterPostContentByAI, payload)
	_, err = t.Client.Enqueue(task)
	if err != nil {
		return false, err
	}
	return true, nil
}
