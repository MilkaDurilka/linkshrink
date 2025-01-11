package controller

import (
	"database/sql"
	"linkshrink/internal/utils/logger"
	"net/http"

	"go.uber.org/zap"
)

type IPingController interface {
	Ping(w http.ResponseWriter, r *http.Request)
}

type PingController struct {
	db     *sql.DB
	logger logger.Logger
}

func NewPingHandler(db *sql.DB, log logger.Logger) *PingController {
	componentLogger := log.With(zap.String("component", "NewPingHandler"))
	return &PingController{db: db, logger: componentLogger}
}

func (c *PingController) Ping(w http.ResponseWriter, r *http.Request) {
	if err := c.db.Ping(); err != nil {
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
