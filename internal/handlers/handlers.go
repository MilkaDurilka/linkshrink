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

func StartServer(
	cfg *config.Config,
	urlController controller.URLController,
	pingController controller.PingController,
	log logger.Logger,
) error {
	r := mux.NewRouter()

	middlewareChain := middleware.InitMiddlewares(log)

	r.Use(middlewareChain)

	r.HandleFunc("/", urlController.ShortenURL).Methods("POST")
	r.HandleFunc("/ping", pingController.Ping).Methods("GET")
	r.HandleFunc("/{id}", urlController.RedirectURL).Methods("GET")
	r.HandleFunc("/api/shorten", urlController.ShortenURLJSON).Methods("POST")
	r.HandleFunc("/api/shorten/batch", urlController.BatchShortenURL).Methods("POST")

	log.Info("Starting server", zap.String("address", cfg.Address))

	err := http.ListenAndServe(cfg.Address, r)

	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
