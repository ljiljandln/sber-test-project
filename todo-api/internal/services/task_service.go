//go:generate mockgen -source=./task_service.go -destination=./mock/task_service.go -package=mock
package services

import (
	"context"

	"todo-api/internal/dto"
	"todo-api/internal/models"
)

type TaskService interface {
	CreateTask(ctx context.Context, req dto.CreateTaskServiceRequest) (*models.Task, error)
	GetTaskByID(ctx context.Context, id uint) (*models.Task, error)
	UpdateTask(ctx context.Context, id uint, req dto.UpdateTaskServiceRequest) (*models.Task, error)
	DeleteTask(ctx context.Context, id uint) error
	ListTasks(ctx context.Context, filter dto.TaskFilter) ([]models.Task, error)
}
