package handlers

import (
	"fmt"
	"linkshrink/internal/config"
	"linkshrink/internal/controller"
	"linkshrink/internal/middleware"
	"linkshrink/internal/utils/logger"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func StartServer(cfg *config.Config, urlController controller.IURLController, log logger.Logger) error {
	r := mux.NewRouter()

	componentLogger := log.With(zap.String("component", "handlers"))

	middlewareChain := middleware.InitMiddlewares(log)

	r.Use(middlewareChain)

	r.HandleFunc("/", urlController.ShortenURL).Methods("POST")
	r.HandleFunc("/{id}", urlController.RedirectURL).Methods("GET")
	r.HandleFunc("/api/shorten", urlController.ShortenURLJSON).Methods("POST")

	componentLogger.Info("Starting server", zap.String("address", cfg.Address))

	err := http.ListenAndServe(cfg.Address, r)

	if err != nil {
		componentLogger.Error("Error on serve", zap.Error(err))
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
