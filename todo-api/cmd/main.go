package main

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"todo-api/internal/controllers"
	"todo-api/internal/models"
	"todo-api/internal/repositories"
	"todo-api/internal/services"
	"todo-api/internal/transport"
	"todo-api/pkg/config"
	"todo-api/pkg/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Config load error details: %+v", err)
		log.Fatal("Config error:", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("DB error:", err)
	}

	if err = db.AutoMigrate(&models.Task{}); err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Migrations completed")

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	repo := repositories.NewTaskRepositoryImpl(db)
	service := services.NewTaskServiceImpl(repo)
	controller := controllers.NewTaskController(service, logger)

	router := transport.SetupRouter(controller, logger)

	if err = router.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
