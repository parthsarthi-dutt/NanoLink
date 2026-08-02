package main

import (
	"log"
	"net/http"
	"os"

	"url-shortener/database"
	"url-shortener/routes"
)

func main() {
	// Initialize Database
	database.InitDB()
	defer database.DB.Close()

	// Initialize Redis
	database.InitRedis()

	// Register Routes
	router := routes.RegisterRoutes()

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
