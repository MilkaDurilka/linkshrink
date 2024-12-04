package main

import (
	"linkshrink/internal/app"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
