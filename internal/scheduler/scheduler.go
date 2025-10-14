package scheduler

import (
	"fmt"
	"log"
	"time"

	"openjobs/connectors/eures"
	"openjobs/pkg/storage"
)

// Scheduler manages periodic job data ingestion
type Scheduler struct {
	euresConnector *eures.EURESConnector
	interval       time.Duration
	stopChan       chan bool
}

// NewScheduler creates a new scheduler instance
func NewScheduler(store *storage.JobStore) *Scheduler {
	return &Scheduler{
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

// runSync executes the job synchronization
func (s *Scheduler) runSync() {
	fmt.Printf("⏰ Running scheduled job sync at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	err := s.euresConnector.SyncJobs()
	if err != nil {
		log.Printf("❌ Scheduled job sync failed: %v", err)
	} else {
		fmt.Println("✅ Scheduled job sync completed successfully")
	}
}

// RunManualSync allows manual triggering of job sync
func (s *Scheduler) RunManualSync() error {
	fmt.Println("🔧 Running manual job sync...")
	return s.euresConnector.SyncJobs()
}
