package main

import (
	"context"
	"linkshrink/internal/app"
	"log"
)

func main() {
	ctx := context.Background()

	if err := app.Run(ctx); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
