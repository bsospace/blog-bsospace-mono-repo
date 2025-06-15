package post

import (
	"encoding/json"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/queue"
	"rag-searchbot-backend/pkg/tiptap"

	"github.com/hibiken/asynq"
)

const TaskTypeFilterPostContentByAI = "ai:filter_post_content"

type FilterPostContentByAIPayload struct {
	Post models.Post
	User models.User
}

type TaskEnqueuer struct {
	QueueRepository queue.QueueRepositoryInterface
	Client          *asynq.Client
}

func NewTaskEnqueuer(client *asynq.Client, QueueRepo queue.QueueRepositoryInterface) *TaskEnqueuer {
	return &TaskEnqueuer{Client: client, QueueRepository: QueueRepo}
}

func (t *TaskEnqueuer) EnqueueFilterPostContentByAI(post *models.Post, user *models.User) (bool, error) {

	plainText := tiptap.ExtractTextFromTiptap(post.Content)

	payload, err := json.Marshal(FilterPostContentByAIPayload{
		Post: models.Post{
			ID:      post.ID,
			Title:   post.Title,
			Content: plainText,
		},
		User: models.User{
			ID:    user.ID,
			Email: user.Email,
		},
	})
	if err != nil {
		return false, err
	}

	// สร้าง Task
	task := asynq.NewTask(TaskTypeFilterPostContentByAI, payload)

	// Enqueue และรับ TaskInfo เพื่อดึง TaskID
	info, err := t.Client.Enqueue(task)
	if err != nil {
		return false, err
	}

	// สร้าง log entry ของ queue พร้อม TaskID
	taskLog := &models.QueueTaskLog{
		TaskID:   info.ID,
		TaskType: TaskTypeFilterPostContentByAI,
		RefID:    post.ID.String(),
		RefType:  "POST",
		Status:   "pending",
		Message:  "",
		Payload:  string(payload),
		UserID:   user.ID,
	}

	if err := t.QueueRepository.Create(taskLog); err != nil {
		return false, err
	}

	return true, nil
}
