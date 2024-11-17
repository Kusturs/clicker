package main

import (
    "log"

    "clicker/internal/app"
    "clicker/internal/config"
)

func main() {
    cfg, err := config.New()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    manager, err := app.NewServerManager(cfg)
    if err != nil {
        log.Fatalf("Failed to create server manager: %v", err)
    }

    if err := manager.Run(); err != nil {
        log.Fatalf("Application error: %v", err)
    }
}
