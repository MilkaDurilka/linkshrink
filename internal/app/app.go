package app

import (
    "linkshrink/internal/controller"
    "linkshrink/internal/repository"
    "linkshrink/internal/service"
    "linkshrink/internal/handlers"

)

func Run() error {
    // Создаем экземпляр репозитория для хранения URL
    urlRepo := repository.NewStore()

    urlService := service.NewURLService(urlRepo)

    urlController := controller.NewURLController(urlService)

    return handlers.StartServer(urlController)
}
