package handlers

import (
	"fmt"
	"linkshrink/internal/config"
	"linkshrink/internal/controller"
	"linkshrink/internal/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func StartServer(cfg *config.Config, urlController controller.IURLController) error {
	r := mux.NewRouter()

	// Создаем логгер
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("cannot create logger: %w", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}() // Отложенная синхронизация логов

	middlewareChain := middleware.InitMiddlewares(logger)
	// Подключаем middleware
	r.Use(middlewareChain)

	r.HandleFunc("/", urlController.ShortenURL).Methods("POST")
	r.HandleFunc("/{id}", urlController.RedirectURL).Methods("GET")
	r.HandleFunc("/api/shorten", urlController.ShortenURLJSON).Methods("POST")

	log.Println("Starting server on: " + cfg.Address)

	err = http.ListenAndServe(cfg.Address, r)

	if err != nil {
		log.Println("Error on serve: ", err)
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
