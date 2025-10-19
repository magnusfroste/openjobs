package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	indeedscraper "openjobs/connectors/indeed-scraper"
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

	// Create Indeed scraper connector
	connector := indeedscraper.NewIndeedScraperConnector(store)

	// Setup HTTP server
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/sync", syncHandler(connector))
	http.HandleFunc("/jobs", jobsHandler(connector))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	fmt.Println("⚠️  ========================================")
	fmt.Println("⚠️  EXPERIMENTAL: Indeed Scraper Plugin")
	fmt.Println("⚠️  Web scraping - Use with caution!")
	fmt.Println("⚠️  Check robots.txt before production use")
	fmt.Println("⚠️  ========================================")
	fmt.Println()
	fmt.Printf("🚀 Indeed Scraper Plugin starting on port %s...\n", port)
	fmt.Printf("📍 Endpoints:\n")
	fmt.Printf("   GET  /health - Health check\n")
	fmt.Printf("   POST /sync   - Trigger scraping sync\n")
	fmt.Printf("   GET  /jobs   - List scraped jobs\n")
	fmt.Println()
	fmt.Println("⏱️  Rate limit: 2 seconds between requests")
	fmt.Println("🔍 Queries: developer, engineer, designer, manager, sales")
	fmt.Println("📄 Pages per query: 3 (30 jobs)")
	fmt.Println("📊 Expected: ~100-120 unique jobs per sync")
	fmt.Println()

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "healthy",
		"connector":    "indeed-scraper",
		"country":      "se",
		"method":       "web_scraping",
		"experimental": true,
		"warning":      "Check robots.txt before production use",
	})
}

func syncHandler(connector *indeedscraper.IndeedScraperConnector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Println("🔄 Scraping sync triggered via HTTP")
		fmt.Println("⚠️  This may take 2-3 minutes due to rate limiting...")
		
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
			"message": "Indeed scraping completed successfully",
			"warning": "Experimental connector - verify data quality",
		})
	}
}

func jobsHandler(connector *indeedscraper.IndeedScraperConnector) http.HandlerFunc {
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
			"success":      true,
			"count":        len(jobs),
			"data":         jobs,
			"experimental": true,
		})
	}
}
