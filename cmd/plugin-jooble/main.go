package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"openjobs/connectors/jooble"
	"openjobs/internal/database"
	"openjobs/pkg/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
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

	// Create Jooble connector
	connector := jooble.NewJoobleConnector(store)

	// Setup HTTP server
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/sync", syncHandler(connector))
	http.HandleFunc("/jobs", jobsHandler(connector))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	fmt.Println("🌐 ========================================")
	fmt.Println("🌐 Jooble Job Aggregator Plugin")
	fmt.Println("🌐 Multi-source job aggregation!")
	fmt.Println("🌐 ========================================")
	fmt.Println()
	fmt.Printf("🚀 Jooble Plugin starting on port %s...\n", port)
	fmt.Printf("📍 Endpoints:\n")
	fmt.Printf("   GET  /health - Health check\n")
	fmt.Printf("   POST /sync   - Trigger job aggregation sync\n")
	fmt.Printf("   GET  /jobs   - List aggregated jobs\n")
	fmt.Println()
	fmt.Println("⏱️  Rate limit: 2 seconds between queries")
	fmt.Println("🔍 Queries: developer, engineer, designer, manager, sales, marketing")
	fmt.Println("📊 Expected: ~200-400 jobs per sync")
	fmt.Println("🌐 Source: Jooble API (aggregates from multiple job boards)")
	fmt.Println()

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "jooble-plugin",
		"version": "1.0.0",
	})
}

func syncHandler(connector *jooble.JoobleConnector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Println("📥 Received sync request for Jooble")

		err := connector.SyncJobs()
		if err != nil {
			log.Printf("Sync failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Jooble sync completed successfully",
		})
	}
}

func jobsHandler(connector *jooble.JoobleConnector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		jobs, err := connector.FetchJobs()
		if err != nil {
			log.Printf("Failed to fetch jobs: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"count":  len(jobs),
			"jobs":   jobs,
		})
	}
}
