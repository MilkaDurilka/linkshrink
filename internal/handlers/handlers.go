package handlers

import (
	"fmt"
	"linkshrink/internal/config"
	"linkshrink/internal/controller"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func StartServer(cfg *config.Config, urlController controller.IURLController) error {
	r := mux.NewRouter()

	r.HandleFunc("/", urlController.ShortenURL).Methods("POST")
	r.HandleFunc("/{id}", urlController.RedirectURL).Methods("GET")

	log.Println("Starting server on: " + cfg.Address)

	err := http.ListenAndServe(cfg.Address, r)

	if err != nil {
		log.Println("Error on serve: ", err)
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
