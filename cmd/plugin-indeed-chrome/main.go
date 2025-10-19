package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	indeedchrome "openjobs/connectors/indeed-chrome"
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

	// Create Indeed Chrome connector
	connector := indeedchrome.NewIndeedChromeConnector(store)

	// Setup HTTP server
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/sync", syncHandler(connector))
	http.HandleFunc("/jobs", jobsHandler(connector))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8087"
	}

	fmt.Println("🌐 ========================================")
	fmt.Println("🌐 Indeed Chrome Scraper Plugin")
	fmt.Println("🌐 Headless Chrome - Bypasses Cloudflare!")
	fmt.Println("🌐 ========================================")
	fmt.Println()
	fmt.Printf("🚀 Indeed Chrome Plugin starting on port %s...\n", port)
	fmt.Printf("📍 Endpoints:\n")
	fmt.Printf("   GET  /health - Health check\n")
	fmt.Printf("   POST /sync   - Trigger Chrome scraping sync\n")
	fmt.Printf("   GET  /jobs   - List scraped jobs\n")
	fmt.Println()
	fmt.Println("⏱️  Rate limit: 3 seconds between requests")
	fmt.Println("🔍 Queries: developer, engineer, designer")
	fmt.Println("📄 Pages per query: 2 (20 jobs)")
	fmt.Println("📊 Expected: ~40-50 unique jobs per sync")
	fmt.Println("🌐 Method: Headless Chrome (bypasses bot detection)")
	fmt.Println()

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"connector": "indeed-chrome",
		"country":   "se",
		"method":    "headless_chrome",
		"advantage": "Bypasses Cloudflare bot detection",
	})
}

func syncHandler(connector *indeedchrome.IndeedChromeConnector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Println("🔄 Chrome scraping sync triggered via HTTP")
		fmt.Println("🌐 This may take 3-5 minutes (Chrome is slower but works!)...")
		
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
			"message": "Indeed Chrome scraping completed successfully",
			"method":  "headless_chrome",
		})
	}
}

func jobsHandler(connector *indeedchrome.IndeedChromeConnector) http.HandlerFunc {
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
			"method":  "headless_chrome",
		})
	}
}
