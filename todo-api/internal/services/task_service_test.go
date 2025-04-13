package services_test

import (
	"context"
	"gorm.io/gorm"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"todo-api/internal/dto"
	"todo-api/internal/models"
	"todo-api/internal/repositories/mock"
	"todo-api/internal/services"
)

func TestTaskService_CreateTask(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		futureDate := time.Now().Add(24 * time.Hour)
		req := dto.CreateTaskServiceRequest{
			Title:       "Test Task",
			Description: "Test Description",
			Date:        futureDate,
		}

		mockRepo.EXPECT().
			Create(
				gomock.Any(),
				gomock.All(
					gomock.AssignableToTypeOf(&models.Task{}),
					gomock.Cond(func(x interface{}) bool {
						task, ok := x.(*models.Task)
						return ok &&
							task.Title == req.Title &&
							task.Description == req.Description &&
							task.Completed == false
					}),
				),
			).
			Return(nil)

		// Act
		task, err := service.CreateTask(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, req.Title, task.Title)
		assert.Equal(t, req.Description, task.Description)
		assert.False(t, task.Completed)
	})

	t.Run("invalid date in past", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		pastDate := time.Now().Add(-24 * time.Hour)
		req := dto.CreateTaskServiceRequest{
			Title: "Test",
			Date:  pastDate,
		}

		// Act
		task, err := service.CreateTask(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "cannot be in the past")
	})
}

func TestTaskService_GetTaskByID(t *testing.T) {
	t.Run("task found", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		expectedTask := &models.Task{
			Model: gorm.Model{
				ID: 1,
			},
			Title: "Test Task",
		}

		mockRepo.EXPECT().
			GetByID(gomock.Any(), uint(1)).
			Return(expectedTask, nil)

		// Act
		task, err := service.GetTaskByID(context.Background(), 1)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedTask, task)
	})

	t.Run("task not found", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		mockRepo.EXPECT().
			GetByID(gomock.Any(), uint(1)).
			Return(nil, assert.AnError)

		// Act
		task, err := service.GetTaskByID(context.Background(), 1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "failed to get task")
	})
}

func TestTaskService_UpdateTask(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		updatedTask := &models.Task{
			Model: gorm.Model{
				ID: 1,
			},
			Title:     "Updated Title",
			Completed: true,
		}

		updatedTitle := "Updated Title"
		updatedCompleted := true

		req := dto.UpdateTaskServiceRequest{
			Title:     &updatedTitle,
			Completed: &updatedCompleted,
		}

		mockRepo.EXPECT().
			Update(gomock.Any(), uint(1), map[string]interface{}{
				"title":     *req.Title,
				"completed": *req.Completed,
			}).
			Return(nil)

		mockRepo.EXPECT().
			GetByID(gomock.Any(), uint(1)).
			Return(updatedTask, nil)

		// Act
		task, err := service.UpdateTask(context.Background(), 1, req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", task.Title)
		assert.True(t, task.Completed)
	})

	t.Run("no updates provided", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		req := dto.UpdateTaskServiceRequest{}

		// Act
		task, err := service.UpdateTask(context.Background(), 1, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "no fields to update")
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		mockRepo.EXPECT().
			Delete(gomock.Any(), uint(1)).
			Return(nil)

		// Act
		err := service.DeleteTask(context.Background(), 1)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		mockRepo.EXPECT().
			Delete(gomock.Any(), uint(1)).
			Return(assert.AnError)

		// Act
		err := service.DeleteTask(context.Background(), 1)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete task")
	})
}

func TestTaskService_ListTasks(t *testing.T) {
	t.Run("with filters", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		completed := true
		expectedTasks := []models.Task{
			{Model: gorm.Model{ID: 1}, Title: "Task 1", Completed: true},
		}

		filter := dto.TaskFilter{
			Completed: &completed,
			Limit:     10,
			Offset:    0,
		}

		mockRepo.EXPECT().
			List(gomock.Any(), filter).
			Return(expectedTasks, nil)

		// Act
		tasks, err := service.ListTasks(context.Background(), filter)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, "Task 1", tasks[0].Title)
	})

	t.Run("default pagination values", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockTaskRepository(ctrl)
		service := services.NewTaskServiceImpl(mockRepo)

		expectedTasks := []models.Task{{Model: gorm.Model{ID: 1}}}

		expectedFilter := dto.TaskFilter{
			Limit:  10,
			Offset: 0,
		}

		mockRepo.EXPECT().
			List(gomock.Any(), expectedFilter).
			Return(expectedTasks, nil)

		// Act
		tasks, err := service.ListTasks(context.Background(), dto.TaskFilter{
			Limit:  -5,
			Offset: -10,
		})

		// Assert
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
	})
}
