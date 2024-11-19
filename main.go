package main

import (
	"archiver/handlers"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "10000" // default port
	}
	// Define new logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Creating new http multiplexer for routing
	router := mux.NewRouter()

	// Routes
	router.HandleFunc("/api/archive/information", handlers.ArchiveInformation).Methods("POST")
	router.HandleFunc("/api/archive/files", handlers.FormArchive).Methods("POST")
	router.HandleFunc("/api/mail/file", handlers.SendEmail).Methods("POST")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	slog.Info("Starting server", "port", port)
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		slog.Error("Failed to start server", "error", err.Error())
	}
}
