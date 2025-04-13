package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"

	"todo-api/internal/dto"
	"todo-api/internal/models"
	"todo-api/internal/services/mock"
)

func setupTestController(t *testing.T) (*TaskController, *mock.MockTaskService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockService := mock.NewMockTaskService(ctrl)
	logger := zaptest.NewLogger(t)
	controller := NewTaskController(mockService, logger)
	return controller, mockService
}

func createTestContext(method string, url string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	} else {
		bodyReader = http.NoBody
	}

	req, _ := http.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	ctx.Request = req

	return ctx, recorder
}

func TestTaskController_CreateTask(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		controller, mockService := setupTestController(t)
		createdTime := "2026-01-02"
		reqBody := dto.CreateTaskRequest{
			Title:       "Test task",
			Description: "Test description",
			DateString:  createdTime,
		}

		parsedDate, err := time.Parse("2006-01-02", createdTime)
		assert.NoError(t, err)

		expectedServiceReq := dto.CreateTaskServiceRequest{
			Title:       reqBody.Title,
			Description: reqBody.Description,
			Date:        parsedDate,
		}

		taskModel := &models.Task{
			Model: gorm.Model{
				ID: 1,
			},
			Title:       reqBody.Title,
			Description: reqBody.Description,
			Date:        parsedDate,
			Completed:   false,
		}

		ctx, recorder := createTestContext("POST", "/tasks/create", reqBody)

		mockService.EXPECT().
			CreateTask(gomock.Any(), expectedServiceReq).
			Return(taskModel, nil).
			Times(1)

		// Act
		controller.CreateTask(ctx)

		// Assert
		assert.Equal(t, http.StatusCreated, recorder.Code)

		var response dto.Response
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "Task created successfully", response.Message)

		dataBytes, err := json.Marshal(response.Data)
		assert.NoError(t, err)

		var taskData models.Task
		err = json.Unmarshal(dataBytes, &taskData)
		assert.NoError(t, err)

		assert.Equal(t, taskModel.ID, taskData.ID)
		assert.Equal(t, taskModel.Title, taskData.Title)
		assert.Equal(t, taskModel.Description, taskData.Description)
		assert.Equal(t, taskModel.Completed, taskData.Completed)
		assert.WithinDuration(t, taskModel.Date, taskData.Date, time.Second) // Проверяем дату
		assert.WithinDuration(t, taskModel.CreatedAt, taskData.CreatedAt, time.Second)
	})

	t.Run("InvalidRequest", func(t *testing.T) {
		// Arrange
		controller, _ := setupTestController(t)
		invalidJSON := `{"title": 123}`
		ctx, recorder := createTestContext("POST", "/tasks/create", invalidJSON)

		// Act
		controller.CreateTask(ctx)

		// Assert
		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Message, "json")
	})

	t.Run("InternalError", func(t *testing.T) {
		// Arrange
		controller, mockService := setupTestController(t)
		reqBody := dto.CreateTaskRequest{
			Title:       "Test task",
			Description: "Test description",
			DateString:  "2026-01-02",
		}

		expectedErr := errors.New("DB connection failed")

		ctx, recorder := createTestContext("POST", "/tasks/create", reqBody)

		mockService.EXPECT().
			CreateTask(gomock.Any(), gomock.Any()).
			Return(nil, expectedErr).
			Times(1)

		// Act
		controller.CreateTask(ctx)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, expectedErr.Error(), response.Message)
	})
}

func TestTaskController_GetTaskByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		controller, mockService := setupTestController(t)
		taskID := uint(1)
		task := &models.Task{
			Model: gorm.Model{
				ID: taskID,
			},
			Title:       "Test task",
			Description: "Test description",
			Completed:   false,
		}

		mockService.EXPECT().
			GetTaskByID(gomock.Any(), taskID).
			Return(task, nil).
			Times(1)

		ctx, recorder := createTestContext("GET", "/tasks/1", nil)
		params := gin.Params{
			{Key: "id", Value: "1"},
		}
		ctx.Params = params

		// Act
		controller.GetTaskByID(ctx)

		// Assert
		assert.Equal(t, http.StatusOK, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "Task retrieved successfully", response.Message)

		dataBytes, err := json.Marshal(response.Data)
		assert.NoError(t, err)

		var taskData models.Task
		err = json.Unmarshal(dataBytes, &taskData)
		assert.NoError(t, err)

		assert.Equal(t, task.ID, taskData.ID)
		assert.Equal(t, task.Title, taskData.Title)
		assert.Equal(t, task.Description, taskData.Description)
		assert.Equal(t, task.Completed, taskData.Completed)
	})

	t.Run("InvalidTaskIDFormat", func(t *testing.T) {
		// Arrange
		controller, _ := setupTestController(t)
		invalidID := "abc"
		ctx, recorder := createTestContext("GET", "/tasks/"+invalidID, nil)
		params := gin.Params{
			{Key: "id", Value: invalidID},
		}
		ctx.Params = params

		// Act
		controller.GetTaskByID(ctx)

		// Assert
		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "Invalid task ID format", response.Message)
	})

	t.Run("ServiceError", func(t *testing.T) {
		// Arrange
		controller, mockService := setupTestController(t)
		taskID := uint(1)
		expectedError := errors.New("database error")

		mockService.EXPECT().
			GetTaskByID(gomock.Any(), taskID).
			Return(nil, expectedError).
			Times(1)

		ctx, recorder := createTestContext("GET", "/tasks/1", nil)
		params := gin.Params{
			{Key: "id", Value: "1"},
		}
		ctx.Params = params

		// Act
		controller.GetTaskByID(ctx)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, expectedError.Error(), response.Message)
		assert.Nil(t, response.Data)
	})
}

func TestTaskController_UpdateTask(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		controller, mockService := setupTestController(t)
		taskID := uint(1)
		updateTime := time.Now()

		title := "Updated title"
		description := "Updated description"
		completed := true
		reqBody := dto.UpdateTaskRequest{
			Title:       &title,
			Description: &description,
			Completed:   &completed,
		}

		expectedServiceReq := dto.UpdateTaskServiceRequest{
			Title:       reqBody.Title,
			Description: reqBody.Description,
			Completed:   reqBody.Completed,
		}

		updatedTask := &models.Task{
			Model: gorm.Model{
				ID:        taskID,
				UpdatedAt: updateTime,
			},
			Title:       *reqBody.Title,
			Description: *reqBody.Description,
			Completed:   *reqBody.Completed,
		}

		ctx, recorder := createTestContext("PUT", "/tasks/1", reqBody)
		params := gin.Params{
			{Key: "id", Value: "1"},
		}
		ctx.Params = params

		mockService.EXPECT().
			UpdateTask(gomock.Any(), taskID, expectedServiceReq).
			Return(updatedTask, nil).
			Times(1)

		// Act
		controller.UpdateTask(ctx)

		// Assert
		assert.Equal(t, http.StatusOK, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "Task updated successfully", response.Message)

		dataBytes, err := json.Marshal(response.Data)
		assert.NoError(t, err)

		var taskData models.Task
		err = json.Unmarshal(dataBytes, &taskData)
		assert.NoError(t, err)

		assert.Equal(t, updatedTask.ID, taskData.ID)
		assert.Equal(t, updatedTask.Title, taskData.Title)
		assert.Equal(t, updatedTask.Description, taskData.Description)
		assert.Equal(t, updatedTask.Completed, taskData.Completed)
	})

	t.Run("InvalidTaskIDFormat", func(t *testing.T) {
		// Arrange
		controller, _ := setupTestController(t)
		invalidID := "abc"
		ctx, recorder := createTestContext("PUT", "/tasks/"+invalidID, nil)
		params := gin.Params{
			{Key: "id", Value: invalidID},
		}
		ctx.Params = params

		// Act
		controller.UpdateTask(ctx)

		// Assert
		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "Invalid task ID format", response.Message)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Arrange
		controller, _ := setupTestController(t)
		invalidJSON := `{"title": 123}`
		ctx, recorder := createTestContext("PUT", "/tasks/1", invalidJSON)
		params := gin.Params{
			{Key: "id", Value: "1"},
		}
		ctx.Params = params

		// Act
		controller.UpdateTask(ctx)

		// Assert
		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Message, "json")
	})

	t.Run("NoFieldsToUpdate", func(t *testing.T) {
		// Arrange
		controller, mockService := setupTestController(t)
		taskID := uint(1)
		reqBody := dto.UpdateTaskRequest{}

		expectedServiceReq := dto.UpdateTaskServiceRequest{
			Title:       nil,
			Description: nil,
			Date:        nil,
			Completed:   nil,
		}

		ctx, recorder := createTestContext("PUT", "/tasks/1", reqBody)
		params := gin.Params{
			{Key: "id", Value: "1"},
		}
		ctx.Params = params

		mockService.EXPECT().
			UpdateTask(gomock.Any(), taskID, expectedServiceReq).
			Return(nil, errors.New("no fields to update")).
			Times(1)

		// Act
		controller.UpdateTask(ctx)

		// Assert
		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "No fields to update", response.Message)
	})

	t.Run("ServiceError", func(t *testing.T) {
		// Arrange
		controller, mockService := setupTestController(t)
		taskID := uint(1)
		title := "Updated title"
		reqBody := dto.UpdateTaskRequest{
			Title: &title,
		}

		expectedServiceReq := dto.UpdateTaskServiceRequest{
			Title:       reqBody.Title,
			Description: nil,
			Date:        nil,
			Completed:   nil,
		}

		expectedError := errors.New("database error")

		ctx, recorder := createTestContext("PUT", "/tasks/1", reqBody)
		params := gin.Params{
			{Key: "id", Value: "1"},
		}
		ctx.Params = params

		mockService.EXPECT().
			UpdateTask(gomock.Any(), taskID, expectedServiceReq).
			Return(nil, expectedError).
			Times(1)

		// Act
		controller.UpdateTask(ctx)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, expectedError.Error(), response.Message)
	})
}

func TestTaskController_DeleteTask(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		controller, mockService := setupTestController(t)
		taskID := uint(1)

		ctx, recorder := createTestContext("DELETE", "/tasks/1", nil)
		params := gin.Params{
			{Key: "id", Value: "1"},
		}
		ctx.Params = params

		mockService.EXPECT().
			DeleteTask(gomock.Any(), taskID).
			Return(nil).
			Times(1)

		// Act
		controller.DeleteTask(ctx)

		// Assert
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Empty(t, recorder.Body.String()) // Проверяем, что тело ответа пустое
	})

	t.Run("InvalidTaskIDFormat", func(t *testing.T) {
		// Arrange
		controller, _ := setupTestController(t)
		invalidID := "abc"
		ctx, recorder := createTestContext("DELETE", "/tasks/"+invalidID, nil)
		params := gin.Params{
			{Key: "id", Value: invalidID},
		}
		ctx.Params = params

		// Act
		controller.DeleteTask(ctx)

		// Assert
		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "Invalid task ID format", response.Message)
	})

	t.Run("ServiceError", func(t *testing.T) {
		// Arrange
		controller, mockService := setupTestController(t)
		taskID := uint(1)
		expectedError := errors.New("database error")

		ctx, recorder := createTestContext("DELETE", "/tasks/1", nil)
		params := gin.Params{
			{Key: "id", Value: "1"},
		}
		ctx.Params = params

		mockService.EXPECT().
			DeleteTask(gomock.Any(), taskID).
			Return(expectedError).
			Times(1)

		// Act
		controller.DeleteTask(ctx)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)

		var response dto.Response
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, expectedError.Error(), response.Message)
	})
}
