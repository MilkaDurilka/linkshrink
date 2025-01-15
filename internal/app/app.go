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
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	urlRepo, err := repository.NewStore(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to connect to repository: %w", err)
	}

	urlService := service.NewURLService(urlRepo)

	urlLogger := logger.With(zap.String("component", "NewURLController"))
	urlController := controller.NewURLController(cfg, urlService, urlLogger)

	pinggLogger := logger.With(zap.String("component", "NewPingHandler"))
	pingController := controller.NewPingHandler(urlRepo, pinggLogger)

	handlersLogger := logger.With(zap.String("component", "handlers"))
	err = handlers.StartServer(cfg, urlController, pingController, handlersLogger)

	if err != nil {
		return fmt.Errorf("failed to start serve: %w", err)
	}

	return nil
}
