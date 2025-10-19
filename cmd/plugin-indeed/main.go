package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"openjobs/connectors/indeed"
	"openjobs/internal/database"
	"openjobs/pkg/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	} else {
		log.Println("✅ Plugin loaded .env file")
	}

	// Connect to shared database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize storage
	store := storage.NewJobStore()

	// Create Indeed connector
	connector := indeed.NewIndeedConnector(store)

	// Setup HTTP server
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/sync", syncHandler(connector))
	http.HandleFunc("/jobs", jobsHandler(connector))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	fmt.Printf("🚀 Indeed Sweden Plugin starting on port %s...\n", port)
	fmt.Printf("📍 Endpoints:\n")
	fmt.Printf("   GET  /health - Health check\n")
	fmt.Printf("   POST /sync   - Trigger job sync\n")
	fmt.Printf("   GET  /jobs   - List synced jobs\n")
	fmt.Println()

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"connector": "indeed",
		"country":   "se",
	})
}

func syncHandler(connector *indeed.IndeedConnector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Println("🔄 Sync triggered via HTTP")
		err := connector.SyncJobs()

		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Indeed jobs synced successfully",
		})
	}
}

func jobsHandler(connector *indeed.IndeedConnector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Fetch jobs from Indeed
		jobs, err := connector.FetchJobs()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"count":   len(jobs),
			"data":    jobs,
		})
	}
}
