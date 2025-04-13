package transport

import (
	"time"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "todo-api/docs"
	"todo-api/internal/controllers"
)

func SetupRouter(taskController *controllers.TaskController, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))

	router.Use(ginzap.RecoveryWithZap(logger, true))

	taskRoutes := router.Group("/tasks")
	{
		taskRoutes.POST("create", taskController.CreateTask)
		taskRoutes.GET("/get/:id", taskController.GetTaskByID)
		taskRoutes.PUT("/update/:id", taskController.UpdateTask)
		taskRoutes.DELETE("/delete/:id", taskController.DeleteTask)
		taskRoutes.GET("list", taskController.ListTasks)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
