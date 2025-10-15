package scheduler

import (
	"fmt"
	"log"
	"os"
	"time"

	"openjobs/connectors/arbetsformedlingen"
	"openjobs/connectors/eures"
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
	afURL := os.Getenv("PLUGIN_ARBETSFORMEDLINGEN_URL")
	euresURL := os.Getenv("PLUGIN_EURES_URL")

	// Use HTTP connectors for external plugins
	if afURL != "" {
		connector := models.NewHTTPPluginConnector("arbetsformedlingen", "Arbetsförmedlingen HTTP Plugin", afURL)
		err := connector.SyncJobs()
		if err != nil {
			log.Printf("❌ Arbetsförmedlingen HTTP sync failed: %v", err)
		} else {
			fmt.Println("✅ Arbetsförmedlingen HTTP sync completed")
		}
	}

	if euresURL != "" {
		connector := models.NewHTTPPluginConnector("eures", "EURES HTTP Plugin", euresURL)
		err := connector.SyncJobs()
		if err != nil {
			log.Printf("❌ EURES HTTP sync failed: %v", err)
		} else {
			fmt.Println("✅ EURES HTTP sync completed")
		}
	}

	// Also try local connectors as fallback (for backward compatibility)
	localConnectors := s.registry.GetEnabledConnectors()
	for _, connector := range localConnectors {
		err := connector.SyncJobs()
		if err != nil {
			log.Printf("❌ %s sync failed: %v", connector.GetName(), err)
		} else {
			fmt.Printf("✅ %s sync completed\n", connector.GetName())
		}
	}

	return nil
}
