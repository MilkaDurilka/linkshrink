package handlers

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "linkshrink/internal/controller"
)


func StartServer(urlController *controller.URLController) error {
    r := mux.NewRouter()

    r.HandleFunc("/", urlController.ShortenURL).Methods("POST")
    r.HandleFunc("/{id}", urlController.RedirectURL).Methods("GET")

    log.Println("Starting server on :8080")
		
    return http.ListenAndServe(":8080", r)
}
