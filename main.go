package main

import (
	"archiver/handlers"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Define new logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Creating new http multiplexer for routing
	router := mux.NewRouter()

	// Routes
	router.HandleFunc("/api/archive/information", handlers.ArchiveInformation).Methods("POST")
	router.HandleFunc("/api/archive/files", handlers.FormArchive).Methods("Post")

	slog.Info("Starting server", "port", "8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		slog.Error("Failed to start server", "error", err.Error())
	}
}
