package jooble

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"openjobs/pkg/models"
	"openjobs/pkg/storage"
)

// JoobleConnector implements connector for Jooble job aggregator
type JoobleConnector struct {
	store      *storage.JobStore
	baseURL    string
	apiKey     string
	userAgent  string
	httpClient *http.Client
}

// JoobleRequest represents the API request structure
type JoobleRequest struct {
	Keywords string `json:"keywords"`
	Location string `json:"location"`
	Page     string `json:"page,omitempty"`
}

// JoobleResponse represents the API response
type JoobleResponse struct {
	TotalCount int         `json:"totalCount"`
	Jobs       []JoobleJob `json:"jobs"`
}

// JoobleJob represents a job from Jooble API
type JoobleJob struct {
	Title       string `json:"title"`
	Location    string `json:"location"`
	Snippet     string `json:"snippet"`
	Salary      string `json:"salary"`
	Source      string `json:"source"`
	Type        string `json:"type"`
	Link        string `json:"link"`
	Company     string `json:"company"`
	Updated     string `json:"updated"`
	ID          int64  `json:"id"` // Changed from string to int64 - Jooble API returns numeric IDs
}

// GetID returns the connector ID
func (jc *JoobleConnector) GetID() string {
	return "jooble"
}

// GetName returns the connector name
func (jc *JoobleConnector) GetName() string {
	return "Jooble Job Aggregator"
}

// NewJoobleConnector creates a new Jooble connector
func NewJoobleConnector(store *storage.JobStore) *JoobleConnector {
	return &JoobleConnector{
		store:      store,
		baseURL:    "https://jooble.org/api",
		apiKey:     os.Getenv("JOOBLE_API_KEY"),
		userAgent:  "OpenJobs-Jooble-Connector/1.0",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchJobs fetches job listings from Jooble API
func (jc *JoobleConnector) FetchJobs() ([]models.JobPost, error) {
	allJobs := []models.JobPost{}

	// If API key not configured, return demo data
	if jc.apiKey == "" {
		fmt.Println("⚠️  JOOBLE_API_KEY not set - returning demo data")
		return jc.getDemoJobs(), nil
	}

	// Get last sync time for incremental sync
	lastSync := jc.getLastSyncTime()

	// Search queries for diverse coverage
	queries := []string{
		"developer",
		"engineer",
		"designer",
		"manager",
		"sales",
		"marketing",
	}

	for _, query := range queries {
		fmt.Printf("🔍 Fetching Jooble jobs for: '%s'\n", query)

		jobs, err := jc.searchJobs(query, "Stockholm")
		if err != nil {
			fmt.Printf("⚠️  Error fetching jobs for '%s': %v\n", query, err)
			continue
		}

		fmt.Printf("   ✅ Found %d jobs for '%s'\n", len(jobs), query)
		allJobs = append(allJobs, jobs...)

		// Rate limiting - be respectful
		time.Sleep(2 * time.Second)
	}

	// Filter by date if we have a last sync time (client-side filtering)
	if !lastSync.IsZero() {
		allJobs = jc.filterJobsByDate(allJobs, lastSync)
		fmt.Printf("📅 Filtered to %d jobs posted after %s\n", len(allJobs), lastSync.Format("2006-01-02"))
	}

	// Deduplicate by ID
	uniqueJobs := jc.deduplicateJobs(allJobs)
	fmt.Printf("📊 Fetched %d unique jobs from Jooble (filtered from %d total)\n", len(uniqueJobs), len(allJobs))

	return uniqueJobs, nil
}

// searchJobs performs a job search via Jooble API
func (jc *JoobleConnector) searchJobs(keywords, location string) ([]models.JobPost, error) {
	url := fmt.Sprintf("%s/%s", jc.baseURL, jc.apiKey)

	// Create request body
	reqBody := JoobleRequest{
		Keywords: keywords,
		Location: location,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", jc.userAgent)

	// Make request
	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jooble API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var joobleResp JoobleResponse
	err = json.Unmarshal(body, &joobleResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Transform to JobPost format
	jobs := make([]models.JobPost, 0, len(joobleResp.Jobs))
	for _, jJob := range joobleResp.Jobs {
		job := jc.transformJoobleJob(jJob)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// transformJoobleJob converts Jooble job format to JobPost
func (jc *JoobleConnector) transformJoobleJob(jj JoobleJob) models.JobPost {
	// Generate unique ID from numeric ID or fallback to link hash
	var jobID string
	if jj.ID != 0 {
		jobID = fmt.Sprintf("%d", jj.ID)
	} else {
		// Fallback: hash of link
		jobID = jc.generateJobID(jj.Link)
	}

	// Parse posted date from updated field
	postedDate := jc.parseJoobleDate(jj.Updated)

	// Clean description
	description := jc.cleanText(jj.Snippet)

	return models.JobPost{
		ID:              fmt.Sprintf("jooble-%s", jobID),
		Title:           strings.TrimSpace(jj.Title),
		Company:         strings.TrimSpace(jj.Company),
		Description:     description,
		Location:        jc.formatLocation(jj.Location),
		Salary:          strings.TrimSpace(jj.Salary),
		SalaryMin:       nil,
		SalaryMax:       nil,
		SalaryCurrency:  "SEK", // Assuming Swedish jobs
		IsRemote:        jc.detectRemote(jj.Title, jj.Snippet, jj.Location),
		URL:             jj.Link,
		EmploymentType:  jc.mapEmploymentType(jj.Type),
		ExperienceLevel: "Mid-level", // Jooble doesn't provide this
		PostedDate:      postedDate,
		ExpiresDate:     postedDate.AddDate(0, 1, 0), // 1 month expiry
		Requirements:    jc.extractRequirements(jj.Title, jj.Snippet),
		Benefits:        []string{},
		Fields: map[string]interface{}{
			"source":         "jooble",
			"source_url":     jj.Link,
			"original_id":    jobID,
			"connector":      "jooble",
			"jooble_source":  jj.Source,
			"jooble_type":    jj.Type,
			"fetched_at":     time.Now(),
		},
	}
}

// generateJobID creates a simple hash from URL
func (jc *JoobleConnector) generateJobID(url string) string {
	// Simple hash: take last part of URL or use timestamp
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// Remove query params
		if idx := strings.Index(lastPart, "?"); idx != -1 {
			lastPart = lastPart[:idx]
		}
		if lastPart != "" {
			return lastPart
		}
	}
	// Fallback to timestamp
	return fmt.Sprintf("%d", time.Now().Unix())
}

// parseJoobleDate parses Jooble date format
func (jc *JoobleConnector) parseJoobleDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now().AddDate(0, 0, -7) // Default to 1 week ago
	}

	// Try common formats
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	// Fallback
	return time.Now().AddDate(0, 0, -7)
}

// mapEmploymentType maps Jooble type to standard format
func (jc *JoobleConnector) mapEmploymentType(jobType string) string {
	jobType = strings.ToLower(jobType)
	
	if strings.Contains(jobType, "full") || strings.Contains(jobType, "heltid") {
		return "Full-time"
	}
	if strings.Contains(jobType, "part") || strings.Contains(jobType, "deltid") {
		return "Part-time"
	}
	if strings.Contains(jobType, "contract") || strings.Contains(jobType, "kontrakt") {
		return "Contract"
	}
	if strings.Contains(jobType, "temporary") || strings.Contains(jobType, "tillfällig") {
		return "Temporary"
	}
	
	return "Full-time" // Default
}

// cleanText removes extra whitespace and cleans text
func (jc *JoobleConnector) cleanText(text string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text = re.ReplaceAllString(text, "")
	
	// Trim whitespace
	text = strings.TrimSpace(text)
	
	// Remove excessive newlines
	re = regexp.MustCompile(`\n{3,}`)
	text = re.ReplaceAllString(text, "\n\n")
	
	return text
}

// getLastSyncTime retrieves the timestamp of the most recent job in database
func (jc *JoobleConnector) getLastSyncTime() time.Time {
	job, err := jc.store.GetMostRecentJob("jooble-")
	if err != nil {
		fmt.Println(" No previous Jooble jobs found - fetching all jobs")
		return time.Time{}
	}
	
	fmt.Printf(" Last Jooble job in database: %s (posted: %s)\n", job.Title, job.PostedDate.Format("2006-01-02"))
	return job.PostedDate
}

// filterJobsByDate filters jobs to only include those posted after the given date
func (jc *JoobleConnector) filterJobsByDate(jobs []models.JobPost, afterDate time.Time) []models.JobPost {
	filtered := make([]models.JobPost, 0, len(jobs))
	
	for _, job := range jobs {
		if job.PostedDate.After(afterDate) || job.PostedDate.Equal(afterDate) {
			filtered = append(filtered, job)
		}
	}
	
	return filtered
}

// formatLocation formats location string
func (jc *JoobleConnector) formatLocation(location string) string {
	location = strings.TrimSpace(location)
	if location == "" {
		return "Sweden"
	}
	
	// Add Sweden if not present
	if !strings.Contains(strings.ToLower(location), "sweden") &&
		!strings.Contains(strings.ToLower(location), "sverige") {
		location = location + ", Sweden"
	}
	
	return location
}

// detectRemote checks if job is remote
func (jc *JoobleConnector) detectRemote(title, description, location string) bool {
	text := strings.ToLower(title + " " + description + " " + location)
	
	remoteKeywords := []string{
		"remote", "distans", "hemarbete", "hemifrån",
		"work from home", "wfh", "anywhere",
	}
	
	for _, keyword := range remoteKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	
	return false
}

// extractRequirements extracts keywords from title and description
func (jc *JoobleConnector) extractRequirements(title, description string) []string {
	requirements := []string{}
	seen := make(map[string]bool)
	
	text := strings.ToLower(title + " " + description)
	
	// Common tech skills
	keywords := []string{
		"Java", "Python", "JavaScript", "TypeScript", "C++", "C#", ".NET", "PHP", "Ruby", "Go", "Rust", "Swift", "Kotlin",
		"React", "Angular", "Vue", "Node.js", "Spring", "Django", "Flask", "Express", "Laravel",
		"Docker", "Kubernetes", "AWS", "Azure", "GCP", "CI/CD", "Jenkins", "Git", "Linux",
		"SQL", "PostgreSQL", "MySQL", "MongoDB", "Redis", "Elasticsearch",
		"API", "REST", "GraphQL", "Microservices", "Agile", "Scrum",
		"Swedish", "English", "B2B", "B2C", "SaaS",
	}
	
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) && !seen[keyword] {
			requirements = append(requirements, keyword)
			seen[keyword] = true
		}
	}
	
	return requirements
}

// deduplicateJobs removes duplicate jobs by ID
func (jc *JoobleConnector) deduplicateJobs(jobs []models.JobPost) []models.JobPost {
	seen := make(map[string]bool)
	unique := []models.JobPost{}
	
	for _, job := range jobs {
		if !seen[job.ID] {
			seen[job.ID] = true
			unique = append(unique, job)
		}
	}
	
	return unique
}

// getDemoJobs returns demo data when API key is not configured
func (jc *JoobleConnector) getDemoJobs() []models.JobPost {
	return []models.JobPost{
		{
			ID:              "jooble-demo-1",
			Title:           "Senior Full Stack Developer",
			Company:         "Tech Company AB",
			Description:     "We are looking for an experienced Full Stack Developer to join our team. You will work with React, Node.js, and PostgreSQL.",
			Location:        "Stockholm, Sweden",
			Salary:          "50000-70000 SEK/month",
			SalaryCurrency:  "SEK",
			IsRemote:        false,
			URL:             "https://jooble.org/demo-job-1",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			PostedDate:      time.Now().Add(-24 * time.Hour),
			ExpiresDate:     time.Now().AddDate(0, 1, 0),
			Requirements:    []string{"React", "Node.js", "PostgreSQL", "JavaScript", "TypeScript"},
			Benefits:        []string{"Remote work options", "Health insurance"},
			Fields: map[string]interface{}{
				"source":        "demo-jooble",
				"source_url":    "https://jooble.org/demo-job-1",
				"original_id":   "demo-1",
				"connector":     "jooble",
				"jooble_source": "Demo",
			},
		},
		{
			ID:              "jooble-demo-2",
			Title:           "DevOps Engineer",
			Company:         "Cloud Solutions AB",
			Description:     "Join our DevOps team! Experience with Kubernetes, Docker, and AWS required.",
			Location:        "Gothenburg, Sweden",
			Salary:          "55000-75000 SEK/month",
			SalaryCurrency:  "SEK",
			IsRemote:        true,
			URL:             "https://jooble.org/demo-job-2",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Mid-level",
			PostedDate:      time.Now().Add(-48 * time.Hour),
			ExpiresDate:     time.Now().AddDate(0, 1, 0),
			Requirements:    []string{"Kubernetes", "Docker", "AWS", "CI/CD", "Linux"},
			Benefits:        []string{"Remote work", "Professional development"},
			Fields: map[string]interface{}{
				"source":        "demo-jooble",
				"source_url":    "https://jooble.org/demo-job-2",
				"original_id":   "demo-2",
				"connector":     "jooble",
				"jooble_source": "Demo",
			},
		},
	}
}

// SyncJobs fetches jobs from Jooble and stores them
func (jc *JoobleConnector) SyncJobs() error {
	startTime := time.Now()
	fmt.Println("🔄 Starting Jooble job aggregator sync...")

	jobs, err := jc.FetchJobs()
	if err != nil {
		// Log failed sync
		jc.store.LogSync(&models.SyncLog{
			ConnectorName:  jc.GetID(),
			StartedAt:      startTime,
			CompletedAt:    time.Now(),
			JobsFetched:    0,
			JobsInserted:   0,
			JobsDuplicates: 0,
			Status:         "failed",
		})
		return fmt.Errorf("failed to fetch jobs from Jooble: %w", err)
	}

	fmt.Printf("📥 Fetched %d jobs from Jooble\n", len(jobs))

	stored := 0
	duplicates := 0
	for _, job := range jobs {
		// Check if job already exists
		existing, err := jc.store.GetJob(job.ID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			fmt.Printf("⚠️  Error checking existing job %s: %v\n", job.ID, err)
			continue
		}

		if existing != nil {
			duplicates++
			continue
		}

		// Store new job
		err = jc.store.CreateJob(&job)
		if err != nil {
			fmt.Printf("❌ Error storing job %s: %v\n", job.ID, err)
			continue
		}

		stored++
		fmt.Printf("✅ Stored job: %s at %s (%s)\n", job.Title, job.Company, job.Location)
	}

	// Log successful sync
	if err := jc.store.LogSync(&models.SyncLog{
		ConnectorName:  jc.GetID(),
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		JobsFetched:    len(jobs),
		JobsInserted:   stored,
		JobsDuplicates: duplicates,
		Status:         "success",
	}); err != nil {
		fmt.Printf("⚠️  Failed to log sync: %v\n", err)
	}

	fmt.Printf("🎉 Jooble sync complete! Fetched: %d, Inserted: %d, Duplicates: %d\n", len(jobs), stored, duplicates)
	return nil
}
