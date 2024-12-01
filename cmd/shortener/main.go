package main

import (
    "log"
    "linkshrink/internal/app"
)

func main() {
    if err := app.Run(); err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
}
