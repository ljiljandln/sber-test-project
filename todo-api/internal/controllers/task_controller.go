package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"todo-api/internal/dto"
	"todo-api/internal/services"
)

type TaskController struct {
	service services.TaskService
	logger  *zap.Logger
}

func NewTaskController(service services.TaskService, logger *zap.Logger) *TaskController {
	return &TaskController{
		service: service,
		logger:  logger,
	}
}

// CreateTask godoc
// @Summary Create a new task
// @Description Add a new task to the system
// @Tags tasks
// @Accept json
// @Produce json
// @Param input body dto.CreateTaskRequest true "Task creation data"
// @Success 201 {object} dto.Response "Task created successfully"
// @Failure 400 {object} dto.Response "Invalid input data"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /tasks/create [post]
func (c *TaskController) CreateTask(ctx *gin.Context) {
	fmt.Println("CreateTask")

	var req dto.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid request data", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	parsedDate, err := time.Parse("2006-01-02", req.DateString)
	if err != nil {
		c.logger.Error("Invalid date format", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid date format, expected YYYY-MM-DD"))
		return
	}

	serviceReq := dto.CreateTaskServiceRequest{
		Title:       req.Title,
		Description: req.Description,
		Date:        parsedDate,
	}

	task, err := c.service.CreateTask(ctx.Request.Context(), serviceReq)
	if err != nil {
		c.logger.Error("Failed to create task", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.logger.Info("Task created successfully", zap.Uint("task_id", task.ID))
	ctx.JSON(http.StatusCreated, dto.SuccessResponse("Task created successfully", task))
}

// GetTaskByID godoc
// @Summary Get task by ID
// @Description Get a single task by its ID
// @Tags tasks
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} dto.Response "Task retrieved successfully"
// @Failure 400 {object} dto.Response "Invalid ID format"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /tasks/get/{id} [get]
func (c *TaskController) GetTaskByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Warn("Invalid task ID format",
			zap.String("id_param", ctx.Param("id")),
			zap.Error(err),
		)
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid task ID format"))
		return
	}

	task, err := c.service.GetTaskByID(ctx.Request.Context(), uint(id))
	if err != nil {
		c.logger.Error("Failed to get task",
			zap.Uint("task_id", uint(id)),
			zap.Error(err),
		)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.logger.Debug("Task retrieved", zap.Uint("task_id", task.ID))
	ctx.JSON(http.StatusOK, dto.SuccessResponse("Task retrieved successfully", task))
}

// UpdateTask godoc
// @Summary Update a task
// @Description Update an existing task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param input body dto.UpdateTaskRequest true "Task update data"
// @Success 200 {object} dto.Response "Task updated successfully"
// @Failure 400 {object} dto.Response "Invalid input data or no fields to update"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /tasks/update/{id} [put]
func (c *TaskController) UpdateTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Warn("Invalid task ID format",
			zap.String("id_param", ctx.Param("id")),
			zap.Error(err),
		)
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid task ID format"))
		return
	}

	var req dto.UpdateTaskRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("Invalid request data", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	serviceReq := dto.UpdateTaskServiceRequest{
		Title:       req.Title,
		Description: req.Description,
		Completed:   req.Completed,
	}

	if req.DateString != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.DateString)
		if err != nil {
			c.logger.Error("Invalid date format", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid date format, expected YYYY-MM-DD"))
			return
		}
		serviceReq.Date = &parsedDate
	}

	task, err := c.service.UpdateTask(ctx.Request.Context(), uint(id), serviceReq)
	if err != nil {
		if err.Error() == "no fields to update" {
			c.logger.Warn("No fields to update provided", zap.Uint("task_id", uint(id)))
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse("No fields to update"))
			return
		}
		c.logger.Error("Failed to update task", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.logger.Info("Task updated successfully", zap.Uint("task_id", task.ID))
	ctx.JSON(http.StatusOK, dto.SuccessResponse("Task updated successfully", task))
}

// DeleteTask godoc
// @Summary Delete a task
// @Description Delete a task by ID
// @Tags tasks
// @Produce json
// @Param id path int true "Task ID"
// @Success 204 "Task deleted successfully"
// @Failure 400 {object} dto.Response "Invalid ID format"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /tasks/delete/{id} [delete]
func (c *TaskController) DeleteTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Warn("Invalid task ID format",
			zap.String("id_param", ctx.Param("id")),
			zap.Error(err),
		)
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid task ID format"))
		return
	}

	if err = c.service.DeleteTask(ctx.Request.Context(), uint(id)); err != nil {
		c.logger.Error("Failed to delete task", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.logger.Info("Task deleted successfully", zap.Uint("task_id", uint(id)))
	ctx.Status(http.StatusOK)
}

// ListTasks godoc
// @Summary List all tasks
// @Description Get a list of tasks with optional filtering
// @Tags tasks
// @Produce json
// @Param completed query bool false "Filter by completion status"
// @Param date_from query string false "Filter by start date (format: 2006-01-02)"
// @Param date_to query string false "Filter by end date (format: 2006-01-02)"
// @Param limit query int false "Limit number of results (default: 10)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} dto.Response "Tasks retrieved successfully"
// @Failure 400 {object} dto.Response "Invalid filter parameters"
// @Failure 500 {object} dto.Response "Internal server error"
// @Router /tasks/list [get]
func (c *TaskController) ListTasks(ctx *gin.Context) {
	filterReq, err := parseTaskFilter(ctx)
	if err != nil {
		c.logger.Error("Invalid filter parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	filter := convertToServiceFilter(filterReq)
	tasks, err := c.service.ListTasks(ctx.Request.Context(), filter)
	if err != nil {
		c.logger.Error("Failed to list tasks", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.logger.Info("Tasks listed successfully", zap.Int("count", len(tasks)))
	ctx.JSON(http.StatusOK, dto.SuccessResponse("Tasks retrieved successfully", tasks))
}

func parseTaskFilter(ctx *gin.Context) (*dto.TaskFilterRequest, error) {
	var filter dto.TaskFilterRequest
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		return nil, err
	}
	return &filter, nil
}

func convertToServiceFilter(req *dto.TaskFilterRequest) dto.TaskFilter {
	filter := dto.TaskFilter{
		Limit:  10,
		Offset: 0,
	}

	if req.Completed != nil {
		if val, err := strconv.ParseBool(*req.Completed); err == nil {
			filter.Completed = &val
		}
	}

	if req.DateFrom != nil {
		if date, err := time.Parse("2006-01-02", *req.DateFrom); err == nil {
			filter.DateFrom = &date
		}
	}

	if req.DateTo != nil {
		if date, err := time.Parse("2006-01-02", *req.DateTo); err == nil {
			filter.DateTo = &date
		}
	}

	if req.Limit != nil && *req.Limit > 0 {
		filter.Limit = *req.Limit
	}

	if req.Offset != nil && *req.Offset >= 0 {
		filter.Offset = *req.Offset
	}

	return filter
}
