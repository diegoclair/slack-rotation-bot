package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/diegoclair/slack-rotation-bot/internal/config"
	"github.com/diegoclair/slack-rotation-bot/internal/database"
	"github.com/diegoclair/slack-rotation-bot/internal/domain/service"
	"github.com/diegoclair/slack-rotation-bot/internal/handlers"
	"github.com/diegoclair/slack-rotation-bot/migrator/sqlite"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg := config.Load()

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Running migrations...")
	if err := sqlite.Migrate(db.DB()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	slackClient := slack.New(cfg.SlackBotToken)

	dataManager := database.NewInstance(db)
	serviceInstance := service.NewInstance(dataManager, slackClient)

	serviceInstance.Scheduler.Start()
	defer serviceInstance.Scheduler.Stop()

	handler := handlers.New(slackClient, serviceInstance.Rotation, cfg.SlackSigningSecret)

	http.HandleFunc("/slack/commands", handler.HandleSlashCommand)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
