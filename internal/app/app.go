package app

import (
	"fmt"
	"linkshrink/internal/config"
	"linkshrink/internal/controller"
	"linkshrink/internal/handlers"
	"linkshrink/internal/repository"
	"linkshrink/internal/service"
	"log"
)

func Run() error {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}

	// Создаем экземпляр репозитория для хранения URL
	urlRepo := repository.NewStore()

	urlService := service.NewURLService(urlRepo)

	urlController := controller.NewURLController(cfg, urlService)

	err = handlers.StartServer(cfg, urlController)

	if err != nil {
		log.Println("Error on start serve", err)
		return fmt.Errorf("failed to start serve: %w", err)
	}

	return nil
}
