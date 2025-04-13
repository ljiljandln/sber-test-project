package repositories_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"todo-api/internal/dto"
	"todo-api/internal/models"
	"todo-api/internal/repositories"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestTaskRepository_Create(t *testing.T) {
	// Arrange
	gormDB, mock := setupMockDB(t)

	repo := repositories.NewTaskRepositoryImpl(gormDB)
	now := time.Now()
	task := &models.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Completed:   false,
		Date:        now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tasks" (.+) VALUES (.+) RETURNING "id"`).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			nil,
			task.Title,
			task.Description,
			task.Date,
			task.Completed,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Act
	err := repo.Create(context.Background(), task)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepository_GetByID(t *testing.T) {
	// Arrange
	gormDB, mock := setupMockDB(t)
	repo := repositories.NewTaskRepositoryImpl(gormDB)

	taskID := uint(1)
	now := time.Now()
	expectedTask := &models.Task{
		Model: gorm.Model{
			ID:        taskID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:       "Test Task",
		Description: "Test Description",
		Completed:   false,
		Date:        now,
	}

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "title", "description", "date", "completed"}).
		AddRow(
			expectedTask.ID,
			expectedTask.CreatedAt,
			expectedTask.UpdatedAt,
			nil, // deleted_at
			expectedTask.Title,
			expectedTask.Description,
			expectedTask.Date,
			expectedTask.Completed,
		)

	expectedSQL := `SELECT * FROM "tasks" WHERE "tasks"."id" = $1 AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`

	mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
		WithArgs(taskID, 1).
		WillReturnRows(rows)

	// Act
	task, err := repo.GetByID(context.Background(), taskID)

	// Assert
	assert.NoError(t, err)

	assert.Equal(t, taskID, task.ID)
	assert.Equal(t, "Test Task", task.Title)
	assert.Equal(t, "Test Description", task.Description)
	assert.Equal(t, false, task.Completed)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepository_Update(t *testing.T) {
	// Arrange
	gormDB, mock := setupMockDB(t)
	repo := repositories.NewTaskRepositoryImpl(gormDB)

	taskID := uint(1)
	updates := map[string]interface{}{
		"title":     "Updated Title",
		"completed": true,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tasks" SET "completed"=$1,"title"=$2,"updated_at"=$3 WHERE id = $4 AND "tasks"."deleted_at" IS NULL`)).
		WithArgs(
			updates["completed"],
			updates["title"],
			sqlmock.AnyArg(), // updated_at
			taskID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err := repo.Update(context.Background(), taskID, updates)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepository_Delete(t *testing.T) {
	// Arrange
	gormDB, mock := setupMockDB(t)
	repo := repositories.NewTaskRepositoryImpl(gormDB)

	taskID := uint(1)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tasks" SET "deleted_at"=$1 WHERE "tasks"."id" = $2 AND "tasks"."deleted_at" IS NULL`)).
		WithArgs(
			sqlmock.AnyArg(), // deleted_at
			taskID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Act
	err := repo.Delete(context.Background(), taskID)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepository_List(t *testing.T) {
	t.Run("with all filters", func(t *testing.T) {
		// Arrange
		gormDB, mock := setupMockDB(t)
		repo := repositories.NewTaskRepositoryImpl(gormDB)

		now := time.Now()
		dateFrom := now.Add(-24 * time.Hour)
		dateTo := now.Add(24 * time.Hour)
		completed := true
		limit := 10
		offset := 0

		filter := dto.TaskFilter{
			Completed: &completed,
			DateFrom:  &dateFrom,
			DateTo:    &dateTo,
			Limit:     limit,
			Offset:    offset,
		}

		expectedTasks := []models.Task{
			{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Title:       "Task 1",
				Description: "Description 1",
				Completed:   true,
				Date:        now,
			},
		}

		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "title", "description", "date", "completed"}).
			AddRow(
				expectedTasks[0].ID,
				expectedTasks[0].CreatedAt,
				expectedTasks[0].UpdatedAt,
				nil,
				expectedTasks[0].Title,
				expectedTasks[0].Description,
				expectedTasks[0].Date,
				expectedTasks[0].Completed,
			)

		expectedSQL := `SELECT * FROM "tasks" WHERE completed = $1 AND (date BETWEEN $2 AND $3) AND "tasks"."deleted_at" IS NULL LIMIT $4`

		mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
			WithArgs(
				*filter.Completed,
				*filter.DateFrom,
				*filter.DateTo,
				filter.Limit,
			).
			WillReturnRows(rows)

		// Act
		tasks, err := repo.List(context.Background(), filter)

		// Assert
		assert.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, expectedTasks[0].ID, tasks[0].ID)
		assert.Equal(t, expectedTasks[0].Title, tasks[0].Title)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with partial filters", func(t *testing.T) {
		// Arrange
		gormDB, mock := setupMockDB(t)
		repo := repositories.NewTaskRepositoryImpl(gormDB)

		now := time.Now()
		dateFrom := now.Add(-24 * time.Hour)
		limit := 5

		filter := dto.TaskFilter{
			DateFrom: &dateFrom,
			Limit:    limit,
		}

		rows := sqlmock.NewRows([]string{"id", "title"}).
			AddRow(1, "Task 1").
			AddRow(2, "Task 2")

		expectedSQL := `SELECT * FROM "tasks" WHERE date >= $1 AND "tasks"."deleted_at" IS NULL LIMIT $2`

		mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
			WithArgs(
				*filter.DateFrom,
				filter.Limit,
			).
			WillReturnRows(rows)

		// Act
		tasks, err := repo.List(context.Background(), filter)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
