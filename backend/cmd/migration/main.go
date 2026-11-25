package main

import (
	"context"
	"flag"
	"log"

	"blog-server/internal/platform/config"
	"blog-server/internal/platform/database"
	logger "blog-server/internal/platform/logging"
	"blog-server/migration"
)

func main() {
	direction := flag.String("direction", "up", "Migration direction: up or down")

	flag.Parse()

	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration loading failed: %s", err)
	}

	logger := logger.NewLogger(cfg)

	ds, err := database.NewDataStore(cfg, logger)
	if err != nil {
		log.Fatalf("database connection failed: %s", err)
	}

	switch *direction {
	case "up":
		if err := migration.Up(ctx, ds.DB()); err != nil {
			log.Fatalf("migration up failed: %s", err)
		}
	case "down":
		if err := migration.Down(ctx, ds.DB()); err != nil {
			log.Fatalf("migration down failed: %s", err)
		}
	}
}
