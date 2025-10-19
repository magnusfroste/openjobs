package scheduler

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"openjobs/connectors/arbetsformedlingen"
	"openjobs/connectors/eures"
	"openjobs/connectors/remoteok"
	"openjobs/connectors/remotive"
	"openjobs/pkg/models"
	"openjobs/pkg/storage"
	
	"github.com/robfig/cron/v3"
)

// Scheduler manages periodic job data ingestion
type Scheduler struct {
	registry     *models.PluginRegistry
	interval     time.Duration
	cronSchedule string
	stopChan     chan bool
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

	// Check for cron schedule first (takes priority)
	cronSchedule := os.Getenv("CRON_SCHEDULE")
	
	// Get sync interval from environment variable (default: 24 hours)
	syncIntervalHours := 24 // Default to once per day
	if envInterval := os.Getenv("SYNC_INTERVAL_HOURS"); envInterval != "" {
		if hours, err := strconv.Atoi(envInterval); err == nil {
			syncIntervalHours = hours
		}
	}
	
	return &Scheduler{
		registry:     registry,
		interval:     time.Hour * time.Duration(syncIntervalHours), // Configurable via SYNC_INTERVAL_HOURS
		cronSchedule: cronSchedule,                                  // Configurable via CRON_SCHEDULE (takes priority)
		stopChan:     make(chan bool),
	}
}

// Start begins the scheduled job ingestion
func (s *Scheduler) Start() {
	// Check if cron schedule is set (takes priority)
	if s.cronSchedule != "" {
		fmt.Printf("⏰ Starting job ingestion with cron schedule: %s\n", s.cronSchedule)
		s.startCronScheduler()
		return
	}
	
	// Otherwise use interval-based scheduling
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

// startCronScheduler starts the cron-based scheduler
func (s *Scheduler) startCronScheduler() {
	c := cron.New()
	
	_, err := c.AddFunc(s.cronSchedule, func() {
		fmt.Printf("\n⏰ Cron triggered at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		s.runSync()
	})
	
	if err != nil {
		log.Fatalf("❌ Invalid cron schedule '%s': %v", s.cronSchedule, err)
	}
	
	c.Start()
	fmt.Printf("✅ Cron scheduler started\n")
	fmt.Printf("📅 Examples:\n")
	fmt.Printf("   '0 6 * * *'   - Every day at 6:00 AM\n")
	fmt.Printf("   '0 */6 * * *' - Every 6 hours\n")
	fmt.Printf("   '0 0 * * *'   - Every day at midnight\n\n")
	
	// Run immediately on start
	go s.runSync()
}

// Stop halts the scheduled job ingestion
func (s *Scheduler) Stop() {
	s.stopChan <- true
}

// runSync executes the job synchronization for all connectors
func (s *Scheduler) runSync() {
	fmt.Printf("⏰ Running scheduled job sync at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// Check if we should use HTTP plugins (microservices mode) or local connectors (monolith mode)
	useHTTPPlugins := os.Getenv("USE_HTTP_PLUGINS") == "true"

	if useHTTPPlugins {
		// Microservices mode: Call plugin containers via HTTP
		fmt.Println("🔌 Using HTTP plugin containers (microservices mode)")
		s.RunManualSync()
	} else {
		// Monolith mode: Run local connectors directly
		fmt.Println("📦 Using local connectors (monolith mode)")
		connectors := s.registry.GetEnabledConnectors()

		for _, connector := range connectors {
			err := connector.SyncJobs()
			if err != nil {
				log.Printf("❌ %s sync failed: %v", connector.GetName(), err)
			} else {
				fmt.Printf("✅ %s sync completed\n", connector.GetName())
			}
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
		"indeed-chrome":      os.Getenv("PLUGIN_INDEED_CHROME_URL"),
		"jooble":             os.Getenv("PLUGIN_JOOBLE_URL"),
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
	if pluginURLs["indeed-chrome"] == "" {
		pluginURLs["indeed-chrome"] = "http://localhost:8087"
	}
	if pluginURLs["jooble"] == "" {
		pluginURLs["jooble"] = "http://localhost:8088"
	}

	// Sync all HTTP plugins
	pluginNames := map[string]string{
		"arbetsformedlingen": "Arbetsförmedlingen",
		"eures":              "EURES",
		"remotive":           "Remotive",
		"remoteok":           "RemoteOK",
		"indeed-chrome":      "Indeed Chrome",
		"jooble":             "Jooble",
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
