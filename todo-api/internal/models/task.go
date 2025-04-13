package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title       string    `gorm:"size:255;not null" json:"title" binding:"required"`
	Description string    `gorm:"type:text" json:"description" binding:"max=1000"`
	Date        time.Time `gorm:"not null;index" json:"date" binding:"required"`
	Completed   bool      `gorm:"default:false" json:"completed"`
}
