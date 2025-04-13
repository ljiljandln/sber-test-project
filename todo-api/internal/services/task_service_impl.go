package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"todo-api/internal/dto"
	"todo-api/internal/models"
	"todo-api/internal/repositories"
)

type TaskServiceImpl struct {
	repo repositories.TaskRepository
}

func NewTaskServiceImpl(repo repositories.TaskRepository) *TaskServiceImpl {
	return &TaskServiceImpl{repo: repo}
}

func (s *TaskServiceImpl) CreateTask(ctx context.Context, req dto.CreateTaskServiceRequest) (*models.Task, error) {
	if req.Date.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, errors.New("task date cannot be in the past")
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Date:        req.Date,
		Completed:   false,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

func (s *TaskServiceImpl) GetTaskByID(ctx context.Context, id uint) (*models.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (s *TaskServiceImpl) UpdateTask(ctx context.Context, id uint, req dto.UpdateTaskServiceRequest) (*models.Task, error) {
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Date != nil {
		updates["date"] = req.Date
	}

	if req.Completed != nil {
		updates["completed"] = *req.Completed
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("error while updating: %w", err)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *TaskServiceImpl) DeleteTask(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (s *TaskServiceImpl) ListTasks(ctx context.Context, filter dto.TaskFilter) ([]models.Task, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	repoFilter := dto.TaskFilter{
		Completed: filter.Completed,
		DateFrom:  filter.DateFrom,
		DateTo:    filter.DateTo,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}

	tasks, err := s.repo.List(ctx, repoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, nil
}
