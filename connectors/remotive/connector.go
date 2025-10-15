package remotive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"openjobs/pkg/models"
	"openjobs/pkg/storage"
)

// RemotiveConnector implements connector for Remotive remote job platform
type RemotiveConnector struct {
	store     *storage.JobStore
	baseURL   string
	userAgent string
}

// RemotiveJob represents a job from the Remotive API
type RemotiveJob struct {
	ID                        int      `json:"id"`
	Title                     string   `json:"title"`
	Description               string   `json:"description"`
	CompanyName               string   `json:"company_name"`
	JobType                   string   `json:"job_type"`
	Salary                    string   `json:"salary"`
	URL                       string   `json:"url"`
	Tags                      []string `json:"tags"`
	CandidateRequiredLocation string   `json:"candidate_required_location"`
	Category                  string   `json:"category"`
	JobType2                  []string `json:"job_type_2"`
	PublicationDate           string   `json:"publication_date"`
}

// RemotiveResponse represents the API response
type RemotiveResponse struct {
	Jobs []RemotiveJob `json:"jobs"`
}

// NewRemotiveConnector creates a new connector
func NewRemotiveConnector(store *storage.JobStore) *RemotiveConnector {
	return &RemotiveConnector{
		store:     store,
		baseURL:   "https://remotive.io/api",
		userAgent: "OpenJobs-Remotive-Connector/1.0",
	}
}

// GetID returns the connector ID
func (rc *RemotiveConnector) GetID() string {
	return "remotive"
}

// GetName returns the connector name
func (rc *RemotiveConnector) GetName() string {
	return "Remotive Remote Jobs Connector"
}

// FetchJobs fetches job listings from Remotive API
func (rc *RemotiveConnector) FetchJobs() ([]models.JobPost, error) {
	url := fmt.Sprintf("%s/remote-jobs?limit=10", rc.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", rc.userAgent)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs from Remotive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("remotive API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var remotiveResponse RemotiveResponse
	err = json.Unmarshal(body, &remotiveResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Transform to our JobPost format
	jobs := make([]models.JobPost, 0, len(remotiveResponse.Jobs))
	for _, remotiveJob := range remotiveResponse.Jobs {
		job := rc.transformRemotiveJob(remotiveJob)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// transformRemotiveJob converts Remotive job format to our JobPost format
func (rc *RemotiveConnector) transformRemotiveJob(rj RemotiveJob) models.JobPost {
	job := models.JobPost{
		ID:              fmt.Sprintf("remotive-%d", rj.ID),
		Title:           rj.Title,
		Company:         rj.CompanyName,
		Description:     rc.extractDescription(rj),
		Location:        rc.formatLocation(rj),
		Salary:          rj.Salary,
		EmploymentType:  rc.mapEmploymentType(rj.JobType),
		ExperienceLevel: "Mid-level", // Most remote jobs are for experienced developers
		PostedDate:      rc.parseRemotiveDate(rj.PublicationDate),
		ExpiresDate:     rc.parseRemotiveDate(rj.PublicationDate).AddDate(0, 2, 0), // 2 month expiration
		Fields: map[string]interface{}{
			"source":                      "remotive",
			"source_url":                  rj.URL,
			"original_id":                 rj.ID,
			"candidate_required_location": rj.CandidateRequiredLocation,
			"category":                    rj.Category,
			"tags":                        rj.Tags,
			"job_type_2":                  rj.JobType2,
			"connector":                   "remotive",
			"fetched_at":                  time.Now(),
		},
	}

	return job
}

// extractDescription uses the title as fallback since Remotive jobs may not have full descriptions
func (rc *RemotiveConnector) extractDescription(rj RemotiveJob) string {
	if rj.Description != "" {
		return rj.Description
	}
	return fmt.Sprintf("Remote %s position at %s", rj.Title, rj.CompanyName)
}

// formatLocation handles remote vs specific location logic
func (rc *RemotiveConnector) formatLocation(rj RemotiveJob) string {
	if rj.CandidateRequiredLocation != "" {
		return rj.CandidateRequiredLocation
	}
	return "Remote"
}

// mapEmploymentType converts Remotive job types to our format
func (rc *RemotiveConnector) mapEmploymentType(remotiveType string) string {
	switch strings.ToLower(remotiveType) {
	case "full_time", "full-time":
		return "Full-time"
	case "part_time", "part-time":
		return "Part-time"
	case "contract", "freelance":
		return "Contract"
	default:
		return "Full-time" // Default for remote jobs
	}
}

// parseRemotiveDate parses Remotive date string (YYYY-MM-DD format)
func (rc *RemotiveConnector) parseRemotiveDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}

	// Try parsing YYYY-MM-DD format
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t
	}

	// Fallback to RFC3339
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t
	}

	return time.Now()
}

// SyncJobs fetches jobs from Remotive and stores them
func (rc *RemotiveConnector) SyncJobs() error {
	fmt.Println("🔄 Starting Remotive remote jobs sync...")

	jobs, err := rc.FetchJobs()
	if err != nil {
		return fmt.Errorf("failed to fetch jobs from Remotive: %w", err)
	}

	fmt.Printf("📥 Fetched %d remote jobs from Remotive\n", len(jobs))

	stored := 0
	for _, job := range jobs {
		// Check if job already exists
		existing, err := rc.store.GetJob(job.ID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			fmt.Printf("⚠️  Error checking existing job %s: %v\n", job.ID, err)
			continue
		}

		if existing != nil {
			// Job already exists, skip
			continue
		}

		// Store new job
		err = rc.store.CreateJob(&job)
		if err != nil {
			fmt.Printf("❌ Error storing job %s: %v\n", job.ID, err)
			continue
		}

		stored++
		fmt.Printf("✅ Stored remote job: %s at %s\n", job.Title, job.Company)
	}

	fmt.Printf("🎉 Remotive sync complete! Stored %d new remote jobs\n", stored)
	return nil
}
