package app

import (
    "linkshrink/internal/controller"
    "linkshrink/internal/repository"
    "linkshrink/internal/service"
    "linkshrink/internal/handlers"
    "linkshrink/internal/config"
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
