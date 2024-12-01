package handlers

import (
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
		
    return http.ListenAndServe(cfg.Address, r)
}
