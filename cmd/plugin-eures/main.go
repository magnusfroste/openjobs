package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"openjobs/connectors/eures"
	"openjobs/internal/database"
	"openjobs/pkg/models"
	"openjobs/pkg/storage"

	"github.com/joho/godotenv"
)

// PluginServer handles HTTP requests for the plugin
type PluginServer struct {
	connector models.PluginConnector
	store     *storage.JobStore
}

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

	store := storage.NewJobStore()
	connector := eures.NewEURESConnector(store)

	server := &PluginServer{
		connector: connector,
		store:     store,
	}

	// Register routes
	http.HandleFunc("/health", server.healthHandler)
	http.HandleFunc("/sync", server.syncHandler)
	http.HandleFunc("/jobs", server.jobsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("EURES Plugin starting on port %s", port)
	log.Printf("Plugin ID: %s", connector.GetID())
	log.Printf("Plugin Name: %s", connector.GetName())

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// healthHandler returns plugin health status
func (s *PluginServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"plugin":    "EURES Connector",
		"plugin_id": s.connector.GetID(),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// syncHandler triggers job synchronization and stores in database
func (s *PluginServer) syncHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("🔄 Starting EURES job sync...")

	err := s.connector.SyncJobs()
	if err != nil {
		log.Printf("❌ Sync failed: %v", err)
		http.Error(w, fmt.Sprintf("Sync failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "EURES sync completed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// jobsHandler returns the latest jobs fetched by this connector
func (s *PluginServer) jobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs, err := s.connector.FetchJobs()
	if err != nil {
		log.Printf("❌ Failed to fetch jobs: %v", err)
		response := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to fetch jobs: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    jobs,
		"count":   len(jobs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
