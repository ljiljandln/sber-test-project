package dto

import (
	"time"
)

type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required,max=255"`
	Description string `json:"description" binding:"max=1000"`
	DateString  string `json:"date" binding:"required"` // "2006-01-02"
}
type CreateTaskServiceRequest struct {
	Title       string
	Description string
	Date        time.Time
}

type UpdateTaskRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=3,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	DateString  *string `json:"date" binding:"omitempty"` // "2006-01-02"
	Completed   *bool   `json:"completed"`
}

type UpdateTaskServiceRequest struct {
	Title       *string
	Description *string
	Date        *time.Time
	Completed   *bool
}

type TaskFilterRequest struct {
	Completed *string `form:"completed"` // "true"/"false"
	DateFrom  *string `form:"date_from"` // "2006-01-02"
	DateTo    *string `form:"date_to"`   // "2006-01-02"
	Limit     *int    `form:"limit"`
	Offset    *int    `form:"offset"`
}

type TaskFilter struct {
	Completed *bool
	DateFrom  *time.Time
	DateTo    *time.Time
	Limit     int
	Offset    int
}
