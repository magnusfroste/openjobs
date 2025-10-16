package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"openjobs/internal/api"
	"openjobs/internal/database"
	"openjobs/internal/scheduler"
	"openjobs/pkg/models"
	"openjobs/pkg/storage"

	"github.com/joho/godotenv"
)

// PluginManager handles plugin lifecycle - DEPRECATED, use PluginRegistry instead
type PluginManager struct {
	plugins map[string]models.PluginInfo
}

// NewPluginManager creates a new plugin manager - DEPRECATED
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]models.PluginInfo),
	}
}

// AddPlugin registers a new plugin - DEPRECATED
func (pm *PluginManager) AddPlugin(plugin models.PluginInfo) {
	pm.plugins[plugin.ID] = plugin
}

// GetPlugin retrieves a plugin by ID - DEPRECATED
func (pm *PluginManager) GetPlugin(id string) (models.PluginInfo, bool) {
	plugin, exists := pm.plugins[id]
	return plugin, exists
}

// GetAllPlugins retrieves all registered plugins - DEPRECATED
func (pm *PluginManager) GetAllPlugins() []models.PluginInfo {
	var plugins []models.PluginInfo
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// UpdatePluginStatus updates plugin status - DEPRECATED
func (pm *PluginManager) UpdatePluginStatus(id, status string) {
	if plugin, exists := pm.GetPlugin(id); exists {
		plugin.Status = status
		plugin.LastRun = time.Now()
		pm.plugins[id] = plugin
	}
}

// Health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := models.APIResponse{
		Success: true,
		Data: map[string]string{
			"status":  "healthy",
			"service": "openjobs",
			"version": "1.0.0",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Plugin registration endpoint
func registerPlugin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var plugin models.PluginInfo
	err := json.NewDecoder(r.Body).Decode(&plugin)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// In a real implementation, this would save to database
	fmt.Printf("Registered plugin: %s (%s)\n", plugin.Name, plugin.ID)

	// DEPRECATED: PluginManager is no longer used; registration is handled by PluginRegistry
	// This endpoint remains for backward compatibility but does not affect job ingestion

	response := models.APIResponse{
		Success: true,
		Data:    plugin,
		Message: "Plugin registered successfully (metadata only, does not affect scheduler)",
	}

	json.NewEncoder(w).Encode(response)
}

// Get all plugins
func getAllPlugins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// DEPRECATED: PluginManager is no longer used; this endpoint returns static mock data
	// In a real implementation, this would fetch enabled plugins from the PluginRegistry
	// Currently, PluginRegistry (in scheduler) is the source of truth for active connectors

	plugins := []models.PluginInfo{
		{
			ID:          "arbetsformedlingen-connector",
			Name:        "Arbetsförmedlingen Connector",
			Version:     "1.0.0",
			Source:      "https://api.arbetsformedlingen.se",
			Status:      "active",
			LastRun:     time.Now().Add(-1 * time.Hour),
			NextRun:     time.Now().Add(1 * time.Hour),
			Description: "Swedish public employment service - government open data",
		},
		{
			ID:          "adzuna-connector",
			Name:        "Adzuna Jobs Connector",
			Version:     "1.0.0",
			Source:      "https://api.adzuna.com",
			Status:      "active",
			LastRun:     time.Now().Add(-2 * time.Hour),
			NextRun:     time.Now().Add(22 * time.Hour),
			Description: "Global job search API with generous free tier",
		},
		{
			ID:          "reed-connector",
			Name:        "Reed.co.uk Connector",
			Version:     "1.0.0",
			Source:      "https://www.reed.co.uk",
			Status:      "active",
			LastRun:     time.Now().Add(-3 * time.Hour),
			NextRun:     time.Now().Add(21 * time.Hour),
			Description: "UK job board with open API access",
		},
		{
			ID:          "eures-connector",
			Name:        "EURES Connector",
			Version:     "1.0.0",
			Source:      "https://eures.europa.eu",
			Status:      "active",
			LastRun:     time.Now().Add(-4 * time.Hour),
			NextRun:     time.Now().Add(20 * time.Hour),
			Description: "European Commission job mobility portal",
		},
	}

	response := models.APIResponse{
		Success: true,
		Data:    plugins,
		Message: "Returns static mock data; active connectors are managed by PluginRegistry in scheduler",
	}

	json.NewEncoder(w).Encode(response)
}

// createSyncHandler creates a handler function for manual sync
func createSyncHandler(server *api.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.SyncJobs(w, r)
	}
}

// createAnalyticsHandler creates a handler function for analytics
func createAnalyticsHandler(server *api.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.AnalyticsHandler(w, r)
	}
}

// Main function to set up routes and start server
func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	} else {
		log.Println("✅ Loaded .env file")
	}

	// Configure and validate Supabase environment
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize storage
	jobStore := storage.NewJobStore()

	// Initialize scheduler for job ingestion
	jobScheduler := scheduler.NewScheduler(jobStore)
	jobScheduler.Start()

	// Initialize API server
	fmt.Printf("🔧 Creating API server with scheduler: %v\n", jobScheduler != nil)
	server := api.NewServer(jobStore, jobScheduler)
	fmt.Printf("✅ API server created: %v\n", server != nil)

	// Set up HTTP routes
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/plugins/register", registerPlugin)
	http.HandleFunc("/plugins", getAllPlugins)

	// Sync routes (must come before /jobs/ to avoid conflicts)
	fmt.Println("📝 Registering route: /sync/manual")
	http.HandleFunc("/sync/manual", createSyncHandler(server))

	// Job routes
	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			server.GetAllJobs(w, r)
		case http.MethodPost:
			server.CreateJob(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Job by ID routes
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			server.GetJobByID(w, r)
		case http.MethodPut:
			server.UpdateJob(w, r)
		case http.MethodDelete:
			server.DeleteJob(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Root redirect to dashboard
	fmt.Println("📝 Registering route: / (redirect to /dashboard)")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	// Dashboard route (use this one)
	fmt.Println("📝 Registering route: /dashboard")
	http.HandleFunc("/dashboard", server.DashboardHandler)

	// Analytics route via helper function
	fmt.Println("📝 Registering route: /analytics")
	http.HandleFunc("/analytics", createAnalyticsHandler(server))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("OpenJobs API starting on port %s\n", port)
	fmt.Printf("🌟 Dashboard available at: http://localhost:%s/dashboard\n", port)

	fmt.Printf("🚀 Server starting... Press Ctrl+C to stop\n")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
