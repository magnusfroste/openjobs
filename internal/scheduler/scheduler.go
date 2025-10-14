package scheduler

import (
	"fmt"
	"log"
	"time"

	"openjobs/connectors/arbetsformedlingen"
	"openjobs/connectors/eures"
	"openjobs/pkg/storage"
)

// Scheduler manages periodic job data ingestion
type Scheduler struct {
	afConnector    *arbetsformedlingen.ArbetsformedlingenConnector
	euresConnector *eures.EURESConnector
	interval       time.Duration
	stopChan       chan bool
}

// NewScheduler creates a new scheduler instance
func NewScheduler(store *storage.JobStore) *Scheduler {
	return &Scheduler{
		afConnector:    arbetsformedlingen.NewArbetsformedlingenConnector(store),
		euresConnector: eures.NewEURESConnector(store),
		interval:       6 * time.Hour, // Run every 6 hours
		stopChan:       make(chan bool),
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

	// Sync Arbetsförmedlingen
	err := s.afConnector.SyncJobs()
	if err != nil {
		log.Printf("❌ Arbetsförmedlingen sync failed: %v", err)
	} else {
		fmt.Println("✅ Arbetsförmedlingen sync completed")
	}

	// Sync EURES
	err = s.euresConnector.SyncJobs()
	if err != nil {
		log.Printf("❌ EURES sync failed: %v", err)
	} else {
		fmt.Println("✅ EURES sync completed")
	}

	fmt.Println("✅ All scheduled syncs completed")
}

// RunManualSync allows manual triggering of job sync for all connectors
func (s *Scheduler) RunManualSync() error {
	fmt.Println("🔧 Running manual job sync for all connectors...")

	// Sync Arbetsförmedlingen
	err := s.afConnector.SyncJobs()
	if err != nil {
		log.Printf("❌ Arbetsförmedlingen sync failed: %v", err)
	}

	// Sync EURES
	err = s.euresConnector.SyncJobs()
	if err != nil {
		log.Printf("❌ EURES sync failed: %v", err)
	}

	return nil // Return nil to indicate overall success (individual errors are logged)
}
