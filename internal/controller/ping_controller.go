package controller

import (
	"linkshrink/internal/repository"
	"linkshrink/internal/utils/logger"
	"net/http"

	"go.uber.org/zap"
)

type IPingController interface {
	Ping(w http.ResponseWriter, r *http.Request)
}

type PingController struct {
	repo   repository.IURLRepository
	logger logger.Logger
}

func NewPingHandler(repo repository.IURLRepository, log logger.Logger) *PingController {
	componentLogger := log.With(zap.String("component", "NewPingHandler"))
	return &PingController{repo: repo, logger: componentLogger}
}

func (c *PingController) Ping(w http.ResponseWriter, r *http.Request) {
	var pingRepo repository.IPingableRepository

	if postgresRepo, ok := c.repo.(repository.IPingableRepository); ok {
		pingRepo = postgresRepo
	} else {
		c.logger.Error("urlRepo does not implement IPingableRepository")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := pingRepo.Ping(); err != nil {
		c.logger.Error("Database ping failed", zap.Error(err))
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status": "OK"}`))
	if err != nil {
		c.logger.Error("Error writing to the response stream", zap.Error(err))
		return
	}
}
