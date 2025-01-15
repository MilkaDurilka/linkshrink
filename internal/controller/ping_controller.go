package controller

import (
	"linkshrink/internal/repository"
	"linkshrink/internal/utils/logger"
	"net/http"

	"go.uber.org/zap"
)

type PingController interface {
	Ping(w http.ResponseWriter, r *http.Request)
}

type PingControllerImpl struct {
	repo   repository.URLRepository
	logger logger.Logger
}

func NewPingHandler(repo repository.URLRepository, log logger.Logger) *PingControllerImpl {
	return &PingControllerImpl{repo: repo, logger: log}
}

func (c *PingControllerImpl) Ping(w http.ResponseWriter, r *http.Request) {
	pingRepo, ok := c.repo.(repository.PingableRepository)

	if !ok {
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
