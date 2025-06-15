package queue

import (
	"rag-searchbot-backend/internal/models"

	"gorm.io/gorm"
)

type QueueRepositoryInterface interface {
	Create(task *models.QueueTaskLog) error
	UpdateStatusByTask(task *models.QueueTaskLog) error
	GetByID(id uint) (*models.QueueTaskLog, error)
	GetByRefID(refID string) ([]*models.QueueTaskLog, error)
	GetByStatus(status string) ([]*models.QueueTaskLog, error)
	GetByTaskType(taskType string) ([]*models.QueueTaskLog, error)
}

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) QueueRepositoryInterface {
	return &Repository{DB: db}
}

func (r *Repository) Create(task *models.QueueTaskLog) error {
	return r.DB.Create(task).Error
}

func (r *Repository) UpdateStatusByTask(task *models.QueueTaskLog) error {
	updates := map[string]any{
		"status":      task.Status,
		"message":     task.Message,
		"started_at":  task.StartedAt,
		"finished_at": task.FinishedAt,
		"duration":    task.Duration,
		"payload":     task.Payload,
		"user_id":     task.UserID,
	}

	return r.DB.Model(&models.QueueTaskLog{}).Where("task_id = ?", task.TaskID).Updates(updates).Error
}

func (r *Repository) GetByID(id uint) (*models.QueueTaskLog, error) {
	var task models.QueueTaskLog
	err := r.DB.First(&task, id).Error
	return &task, err
}

func (r *Repository) GetByRefID(refID string) ([]*models.QueueTaskLog, error) {
	var tasks []*models.QueueTaskLog
	err := r.DB.Where("ref_id = ?", refID).Find(&tasks).Error
	return tasks, err
}

func (r *Repository) GetByStatus(status string) ([]*models.QueueTaskLog, error) {
	var tasks []*models.QueueTaskLog
	err := r.DB.Where("status = ?", status).Find(&tasks).Error
	return tasks, err
}

func (r *Repository) GetByTaskType(taskType string) ([]*models.QueueTaskLog, error) {
	var tasks []*models.QueueTaskLog
	err := r.DB.Where("task_type = ?", taskType).Find(&tasks).Error
	return tasks, err
}
