package ai

import (
	"encoding/json"
	"rag-searchbot-backend/internal/models"

	"github.com/hibiken/asynq"
)

const TaskTypeEmbedPost = "ai:embed_post"

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
			ID:    post.ID,
			Title: post.Title,
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
