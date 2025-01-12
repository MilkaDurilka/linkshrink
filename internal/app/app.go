package app

import (
	"fmt"
	"linkshrink/internal/config"
	"linkshrink/internal/controller"
	"linkshrink/internal/handlers"
	"linkshrink/internal/repository"
	"linkshrink/internal/service"

	"go.uber.org/zap"
)

func Run() error {
	// Создаем логгер
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("cannot create logger: %w", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}() // Отложенная синхронизация логов

	cfg, err := config.InitConfig()
	if err != nil {
		logger.Error("Error initializing config", zap.Error(err))
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	urlRepo, err := repository.NewStore(cfg, logger)
	if err != nil {
		logger.Error("Unable to connect to repository", zap.Error(err))
		return fmt.Errorf("failed to connect to repository: %w", err)
	}

	urlService := service.NewURLService(urlRepo)

	urlController := controller.NewURLController(cfg, urlService, logger)

	pingController := controller.NewPingHandler(urlRepo, logger)

	err = handlers.StartServer(cfg, urlController, pingController, logger)

	if err != nil {
		logger.Error("Error on start serve", zap.Error(err))
		return fmt.Errorf("failed to start serve: %w", err)
	}

	return nil
}
