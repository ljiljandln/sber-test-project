//go:generate mockgen -source=./task_repository.go -destination=./mock/task_repository.go -package=mock
package repositories

import (
	"context"

	"todo-api/internal/dto"
	"todo-api/internal/models"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id uint) (*models.Task, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter dto.TaskFilter) ([]models.Task, error)
}
