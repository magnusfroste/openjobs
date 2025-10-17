package scheduler

import (
	"fmt"
	"log"
	"os"
	"time"

	"openjobs/connectors/arbetsformedlingen"
	"openjobs/connectors/eures"
	"openjobs/connectors/remoteok"
	"openjobs/connectors/remotive"
	"openjobs/pkg/models"
	"openjobs/pkg/storage"
)

// Scheduler manages periodic job data ingestion
type Scheduler struct {
	registry *models.PluginRegistry
	interval time.Duration
	stopChan chan bool
}

// NewScheduler creates a new scheduler instance
func NewScheduler(store *storage.JobStore) *Scheduler {
	// Create plugin registry
	registry := models.NewPluginRegistry()

	// Register built-in connectors
	registry.Register(arbetsformedlingen.NewArbetsformedlingenConnector(store))
	registry.Register(eures.NewEURESConnector(store))
	registry.Register(remoteok.NewRemoteOKConnector(store))
	registry.Register(remotive.NewRemotiveConnector(store))

	return &Scheduler{
		registry: registry,
		interval: time.Hour * 6, // Run every 6 hours
		stopChan: make(chan bool),
	}
}

// Start begins the scheduled job ingestion
func (s *Scheduler) Start() {
	fmt.Printf("🚀 Starting job ingestion scheduler (every %v)\n", s.interval)

	// Run immediately on start
	go s.runSync()

	// Then run on schedule
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.runSync()
			case <-s.stopChan:
				ticker.Stop()
				fmt.Println("🛑 Job ingestion scheduler stopped")
				return
			}
		}
	}()
}

// Stop halts the scheduled job ingestion
func (s *Scheduler) Stop() {
	s.stopChan <- true
}

// runSync executes the job synchronization for all connectors
func (s *Scheduler) runSync() {
	fmt.Printf("⏰ Running scheduled job sync at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// Get all enabled connectors from registry
	connectors := s.registry.GetEnabledConnectors()

	for _, connector := range connectors {
		// Check if connector is enabled (future feature)
		// if !connector.IsEnabled() { continue }

		err := connector.SyncJobs()
		if err != nil {
			log.Printf("❌ %s sync failed: %v", connector.GetName(), err)
		} else {
			fmt.Printf("✅ %s sync completed\n", connector.GetName())
		}
	}

	fmt.Println("✅ All scheduled syncs completed")
}

// RunManualSync allows manual triggering of job sync for all connectors
func (s *Scheduler) RunManualSync() error {
	fmt.Println("🔧 Running manual job sync for all connectors...")

	// Check environment variables for external plugin URLs
	pluginURLs := map[string]string{
		"arbetsformedlingen": os.Getenv("PLUGIN_ARBETSFORMEDLINGEN_URL"),
		"eures":              os.Getenv("PLUGIN_EURES_URL"),
		"remotive":           os.Getenv("PLUGIN_REMOTIVE_URL"),
		"remoteok":           os.Getenv("PLUGIN_REMOTEOK_URL"),
	}

	// Default URLs for Docker Compose setup
	if pluginURLs["arbetsformedlingen"] == "" {
		pluginURLs["arbetsformedlingen"] = "http://localhost:8081"
	}
	if pluginURLs["eures"] == "" {
		pluginURLs["eures"] = "http://localhost:8082"
	}
	if pluginURLs["remotive"] == "" {
		pluginURLs["remotive"] = "http://localhost:8083"
	}
	if pluginURLs["remoteok"] == "" {
		pluginURLs["remoteok"] = "http://localhost:8084"
	}

	// Sync all HTTP plugins
	pluginNames := map[string]string{
		"arbetsformedlingen": "Arbetsförmedlingen",
		"eures":              "EURES",
		"remotive":           "Remotive",
		"remoteok":           "RemoteOK",
	}

	for id, url := range pluginURLs {
		if url != "" {
			name := pluginNames[id]
			connector := models.NewHTTPPluginConnector(id, name+" HTTP Plugin", url)
			err := connector.SyncJobs()
			if err != nil {
				log.Printf("❌ %s HTTP sync failed: %v", name, err)
			} else {
				fmt.Printf("✅ %s HTTP sync completed\n", name)
			}
		}
	}

	// NOTE: Do NOT run local connectors here - they are already running as HTTP plugins
	// Running both would cause duplicate sync logs and duplicate job entries
	// The local connectors in the registry are only used for scheduled syncs in non-microservice mode

	return nil
}
