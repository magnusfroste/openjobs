package remoteok

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

// RemoteOKConnector implements connector for RemoteOK.com
type RemoteOKConnector struct {
	store     *storage.JobStore
	baseURL   string
	userAgent string
}

// RemoteOKJob represents a job from the RemoteOK API
type RemoteOKJob struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Position    string   `json:"position"`
	Company     string   `json:"company"`
	CompanyLogo string   `json:"company_logo"`
	Location    string   `json:"location"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Date        string   `json:"date"`
	URL         string   `json:"url"`
	ApplyURL    string   `json:"apply_url"`
}

// NewRemoteOKConnector creates a new connector
func NewRemoteOKConnector(store *storage.JobStore) *RemoteOKConnector {
	return &RemoteOKConnector{
		store:     store,
		baseURL:   "https://remoteok.com/api",
		userAgent: "OpenJobs-RemoteOK-Connector/1.0",
	}
}

// GetID returns the connector ID
func (rc *RemoteOKConnector) GetID() string {
	return "remoteok"
}

// GetName returns the connector name
func (rc *RemoteOKConnector) GetName() string {
	return "RemoteOK Connector"
}

// FetchJobs fetches job listings from RemoteOK API
func (rc *RemoteOKConnector) FetchJobs() ([]models.JobPost, error) {
	url := rc.baseURL

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", rc.userAgent)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs from RemoteOK: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("remoteOK API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var remoteOKJobs []RemoteOKJob
	err = json.Unmarshal(body, &remoteOKJobs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// First item is metadata, skip it
	if len(remoteOKJobs) > 0 {
		remoteOKJobs = remoteOKJobs[1:]
	}

	// Get last sync time for incremental sync
	lastSync := rc.getLastSyncTime()
	
	// Filter jobs to only new ones (posted after last sync)
	filteredJobs := []RemoteOKJob{}
	for _, remoteOKJob := range remoteOKJobs {
		jobDate := rc.parseRemoteOKDate(remoteOKJob.Date)
		if lastSync.IsZero() || jobDate.After(lastSync) {
			filteredJobs = append(filteredJobs, remoteOKJob)
		}
	}
	
	fmt.Printf("📊 Filtered %d jobs from %d total (only new jobs)\n", len(filteredJobs), len(remoteOKJobs))

	// Transform to our JobPost format
	jobs := make([]models.JobPost, 0, len(filteredJobs))
	for _, remoteOKJob := range filteredJobs {
		job := rc.transformRemoteOKJob(remoteOKJob)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// transformRemoteOKJob converts RemoteOK job format to our JobPost format
func (rc *RemoteOKConnector) transformRemoteOKJob(rj RemoteOKJob) models.JobPost {
	// Extract URL
	url := rc.extractURL(rj)
	
	job := models.JobPost{
		ID:              fmt.Sprintf("remoteok-%s", rj.ID),
		Title:           rj.Position,
		Company:         rj.Company,
		Description:     rc.extractDescription(rj),
		Location:        rc.formatLocation(rj),
		Salary:          "", // RemoteOK doesn't provide salary
		SalaryMin:       nil,
		SalaryMax:       nil,
		SalaryCurrency:  "USD",
		IsRemote:        true, // ⭐ All RemoteOK jobs are remote
		URL:             url,  // ⭐ Direct application URL
		EmploymentType:  "Full-time",
		ExperienceLevel: "Mid-level", // Most remote jobs are for experienced developers
		PostedDate:      rc.parseRemoteOKDate(rj.Date),
		ExpiresDate:     rc.parseRemoteOKDate(rj.Date).AddDate(0, 2, 0), // 2 month expiration
		Requirements:    rc.extractRequirements(rj), // Tags + keyword extraction
		Benefits:        []string{"Remote work"},
		Fields: map[string]interface{}{
			"source":       "remoteok",
			"source_url":   url,
			"original_id":  rj.ID,
			"slug":         rj.Slug,
			"tags":         rj.Tags,
			"company_logo": rj.CompanyLogo,
			"apply_url":    rj.ApplyURL,
			"connector":    "remoteok",
			"fetched_at":   time.Now(),
		},
	}

	return job
}

// extractDescription uses the description or creates a simple one
func (rc *RemoteOKConnector) extractDescription(rj RemoteOKJob) string {
	if rj.Description != "" {
		return rj.Description
	}
	return fmt.Sprintf("Remote %s position at %s", rj.Position, rj.Company)
}

// formatLocation handles remote location
func (rc *RemoteOKConnector) formatLocation(rj RemoteOKJob) string {
	if rj.Location != "" && rj.Location != "Remote" {
		return rj.Location + " (Remote)"
	}
	return "Remote"
}

// extractURL gets the job URL
func (rc *RemoteOKConnector) extractURL(rj RemoteOKJob) string {
	if rj.URL != "" {
		return rj.URL
	}
	if rj.Slug != "" {
		return fmt.Sprintf("https://remoteok.com/remote-jobs/%s", rj.Slug)
	}
	return fmt.Sprintf("https://remoteok.com/remote-jobs/%s", rj.ID)
}

// parseRemoteOKDate parses RemoteOK date string (Unix timestamp or ISO format)
func (rc *RemoteOKConnector) parseRemoteOKDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}

	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t
	}

	// Try parsing as YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t
	}

	return time.Now()
}

// SyncJobs fetches jobs from RemoteOK and stores them
func (rc *RemoteOKConnector) SyncJobs() error {
	startTime := time.Now()
	fmt.Println("🔄 Starting RemoteOK remote jobs sync...")

	jobs, err := rc.FetchJobs()
	if err != nil {
		// Log failed sync
		rc.store.LogSync(&models.SyncLog{
			ConnectorName: rc.GetID(),
			StartedAt:     startTime,
			CompletedAt:   time.Now(),
			JobsFetched:   0,
			JobsInserted:  0,
			JobsDuplicates: 0,
			Status:        "failed",
		})
		return fmt.Errorf("failed to fetch jobs from RemoteOK: %w", err)
	}

	fmt.Printf("📥 Fetched %d remote jobs from RemoteOK\n", len(jobs))

	stored := 0
	duplicates := 0
	for _, job := range jobs {
		// Check if job already exists
		existing, err := rc.store.GetJob(job.ID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			fmt.Printf("⚠️  Error checking existing job %s: %v\n", job.ID, err)
			continue
		}

		if existing != nil {
			// Job already exists, skip
			duplicates++
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

	// Log successful sync
	if err := rc.store.LogSync(&models.SyncLog{
		ConnectorName:  rc.GetID(),
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		JobsFetched:    len(jobs),
		JobsInserted:   stored,
		JobsDuplicates: duplicates,
		Status:         "success",
	}); err != nil {
		fmt.Printf("⚠️  Failed to log sync: %v\n", err)
	}

	fmt.Printf("🎉 RemoteOK sync complete! Fetched: %d, Inserted: %d, Duplicates: %d\n", len(jobs), stored, duplicates)
	return nil
}

// extractRequirements extracts keywords from tags, title, and description
func (rc *RemoteOKConnector) extractRequirements(rj RemoteOKJob) []string {
requirements := []string{}
seen := make(map[string]bool)

// Add tags first (most reliable)
for _, tag := range rj.Tags {
if tag != "" && !seen[tag] {
requirements = append(requirements, tag)
seen[tag] = true
}
}

// Extract from title and description
text := strings.ToLower(rj.Position + " " + rj.Description)

// Common tech skills
keywords := []string{
"Java", "Python", "JavaScript", "TypeScript", "C++", "C#", ".NET", "PHP", "Ruby", "Go", "Rust", "Swift", "Kotlin",
"React", "Angular", "Vue", "Node.js", "Spring", "Django", "Flask", "Express", "Laravel",
"Docker", "Kubernetes", "AWS", "Azure", "GCP", "CI/CD", "Jenkins", "Git", "Linux",
"SQL", "PostgreSQL", "MySQL", "MongoDB", "Redis", "Elasticsearch",
"API", "REST", "GraphQL", "Microservices", "Agile", "Scrum",
}

for _, keyword := range keywords {
if strings.Contains(text, strings.ToLower(keyword)) && !seen[keyword] {
requirements = append(requirements, keyword)
seen[keyword] = true
}
}

return requirements
}

// getLastSyncTime retrieves the timestamp of the most recent job in database
func (rc *RemoteOKConnector) getLastSyncTime() time.Time {
	job, err := rc.store.GetMostRecentJob("remoteok-")
	if err != nil {
		fmt.Println("📅 No previous RemoteOK jobs found - processing all jobs")
		return time.Time{}
	}
	
	fmt.Printf("📅 Last RemoteOK job in database: %s (posted: %s)\n", job.Title, job.PostedDate.Format("2006-01-02"))
	return job.PostedDate
}
