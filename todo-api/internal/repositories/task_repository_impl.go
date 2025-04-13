package repositories

import (
	"context"

	"gorm.io/gorm"

	"todo-api/internal/dto"
	"todo-api/internal/models"
)

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepositoryImpl(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) GetByID(ctx context.Context, id uint) (*models.Task, error) {
	var task models.Task
	err := r.db.WithContext(ctx).First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *taskRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Task{}, id).Error
}

func (r *taskRepository) List(ctx context.Context, filter dto.TaskFilter) ([]models.Task, error) {
	var tasks []models.Task
	query := r.db.WithContext(ctx).Model(&models.Task{})

	if filter.Completed != nil {
		query = query.Where("completed = ?", *filter.Completed)
	}

	if filter.DateFrom != nil || filter.DateTo != nil {
		if filter.DateFrom != nil && filter.DateTo != nil {
			query = query.Where("date BETWEEN ? AND ?", *filter.DateFrom, *filter.DateTo)
		} else if filter.DateFrom != nil {
			query = query.Where("date >= ?", *filter.DateFrom)
		} else {
			query = query.Where("date <= ?", *filter.DateTo)
		}
	}

	err := query.Limit(filter.Limit).Offset(filter.Offset).Find(&tasks).Error
	return tasks, err
}
