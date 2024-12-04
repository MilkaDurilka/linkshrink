package app

import (
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

	return handlers.StartServer(cfg, urlController)
}
