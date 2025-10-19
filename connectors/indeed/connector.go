package indeed

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"openjobs/pkg/models"
	"openjobs/pkg/storage"
)

// IndeedConnector implements connector for Indeed.se (Sweden)
type IndeedConnector struct {
	store       *storage.JobStore
	baseURL     string
	publisherID string
	userAgent   string
	country     string // "se" for Sweden
}

// IndeedResponse represents the API response from Indeed
type IndeedResponse struct {
	Version       int          `json:"version"`
	Query         string       `json:"query"`
	Location      string       `json:"location"`
	TotalResults  int          `json:"totalResults"`
	Start         int          `json:"start"`
	End           int          `json:"end"`
	PageNumber    int          `json:"pageNumber"`
	Results       []IndeedJob  `json:"results"`
}

// IndeedJob represents a job from the Indeed API
type IndeedJob struct {
	JobTitle             string    `json:"jobtitle"`
	Company              string    `json:"company"`
	City                 string    `json:"city"`
	State                string    `json:"state"`
	Country              string    `json:"country"`
	FormattedLocation    string    `json:"formattedLocation"`
	Source               string    `json:"source"`
	Date                 string    `json:"date"`
	Snippet              string    `json:"snippet"`
	URL                  string    `json:"url"`
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	JobKey               string    `json:"jobkey"`
	Sponsored            bool      `json:"sponsored"`
	Expired              bool      `json:"expired"`
	FormattedRelativeTime string   `json:"formattedRelativeTime"`
}

// NewIndeedConnector creates a new Indeed connector
func NewIndeedConnector(store *storage.JobStore) *IndeedConnector {
	publisherID := os.Getenv("INDEED_PUBLISHER_ID")
	if publisherID == "" {
		publisherID = "demo" // Demo mode for testing
		fmt.Println("⚠️  INDEED_PUBLISHER_ID not set, using demo mode (limited results)")
	}

	return &IndeedConnector{
		store:       store,
		baseURL:     "http://api.indeed.com/ads/apisearch",
		publisherID: publisherID,
		userAgent:   "OpenJobs-Indeed-Connector/1.0",
		country:     "se", // Sweden
	}
}

// GetID returns the connector ID
func (ic *IndeedConnector) GetID() string {
	return "indeed"
}

// GetName returns the connector name
func (ic *IndeedConnector) GetName() string {
	return "Indeed Sweden Connector"
}

// FetchJobs fetches job listings from Indeed API
func (ic *IndeedConnector) FetchJobs() ([]models.JobPost, error) {
	allJobs := []models.JobPost{}
	
	// Search queries to get diverse jobs
	queries := []string{
		"",                    // All jobs
		"developer",           // Tech jobs
		"engineer",            // Engineering jobs
		"manager",             // Management jobs
		"sales",               // Sales jobs
		"customer service",    // Service jobs
	}

	// Get last sync time for incremental sync
	lastSync := ic.getLastSyncTime()
	
	for _, query := range queries {
		fmt.Printf("🔍 Searching Indeed for: '%s'\n", query)
		
		// Fetch multiple pages (up to 100 results per query)
		for start := 0; start < 100; start += 25 {
			jobs, err := ic.fetchJobsPage(query, start)
			if err != nil {
				fmt.Printf("⚠️  Error fetching page %d for query '%s': %v\n", start/25+1, query, err)
				continue
			}
			
			if len(jobs) == 0 {
				break // No more results
			}
			
			// Filter to only new jobs
			for _, job := range jobs {
				if lastSync.IsZero() || job.PostedDate.After(lastSync) {
					allJobs = append(allJobs, job)
				}
			}
			
			// Rate limiting - be nice to Indeed API
			time.Sleep(1 * time.Second)
		}
	}
	
	// Deduplicate by job key
	uniqueJobs := ic.deduplicateJobs(allJobs)
	
	fmt.Printf("📊 Fetched %d unique jobs from Indeed (filtered from %d total)\n", len(uniqueJobs), len(allJobs))
	
	return uniqueJobs, nil
}

// fetchJobsPage fetches a single page of jobs
func (ic *IndeedConnector) fetchJobsPage(query string, start int) ([]models.JobPost, error) {
	// Build API URL
	params := url.Values{}
	params.Set("publisher", ic.publisherID)
	params.Set("v", "2")
	params.Set("format", "json")
	params.Set("co", ic.country)
	params.Set("limit", "25")
	params.Set("start", fmt.Sprintf("%d", start))
	
	if query != "" {
		params.Set("q", query)
	}
	
	// Add user IP and user agent if available
	params.Set("useragent", ic.userAgent)
	
	apiURL := fmt.Sprintf("%s?%s", ic.baseURL, params.Encode())
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", ic.userAgent)
	req.Header.Set("Accept", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs from Indeed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("indeed API error %d: %s", resp.StatusCode, string(body))
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	var indeedResp IndeedResponse
	err = json.Unmarshal(body, &indeedResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Transform to our JobPost format
	jobs := make([]models.JobPost, 0, len(indeedResp.Results))
	for _, indeedJob := range indeedResp.Results {
		if indeedJob.Expired {
			continue // Skip expired jobs
		}
		job := ic.transformIndeedJob(indeedJob)
		jobs = append(jobs, job)
	}
	
	return jobs, nil
}

// transformIndeedJob converts Indeed job format to our JobPost format
func (ic *IndeedConnector) transformIndeedJob(ij IndeedJob) models.JobPost {
	job := models.JobPost{
		ID:              fmt.Sprintf("indeed-%s", ij.JobKey),
		Title:           ij.JobTitle,
		Company:         ij.Company,
		Description:     ic.cleanSnippet(ij.Snippet),
		Location:        ic.formatLocation(ij),
		Salary:          "", // Indeed doesn't provide salary in API
		SalaryMin:       nil,
		SalaryMax:       nil,
		SalaryCurrency:  "SEK", // Sweden
		IsRemote:        ic.detectRemote(ij),
		URL:             ij.URL,
		EmploymentType:  "Full-time", // Indeed doesn't specify in API
		ExperienceLevel: "Mid-level",
		PostedDate:      ic.parseIndeedDate(ij.Date),
		ExpiresDate:     ic.parseIndeedDate(ij.Date).AddDate(0, 1, 0), // 1 month expiration
		Requirements:    ic.extractRequirements(ij),
		Benefits:        []string{},
		Fields: map[string]interface{}{
			"source":                  "indeed",
			"source_url":              ij.URL,
			"original_id":             ij.JobKey,
			"city":                    ij.City,
			"state":                   ij.State,
			"country":                 ij.Country,
			"formatted_location":      ij.FormattedLocation,
			"sponsored":               ij.Sponsored,
			"formatted_relative_time": ij.FormattedRelativeTime,
			"latitude":                ij.Latitude,
			"longitude":               ij.Longitude,
			"connector":               "indeed",
			"fetched_at":              time.Now(),
		},
	}
	
	return job
}

// cleanSnippet removes HTML tags and cleans up the snippet
func (ic *IndeedConnector) cleanSnippet(snippet string) string {
	// Remove <b> tags (Indeed uses them for highlighting)
	snippet = strings.ReplaceAll(snippet, "<b>", "")
	snippet = strings.ReplaceAll(snippet, "</b>", "")
	
	// Trim whitespace
	snippet = strings.TrimSpace(snippet)
	
	return snippet
}

// formatLocation formats the location string
func (ic *IndeedConnector) formatLocation(ij IndeedJob) string {
	if ij.FormattedLocation != "" {
		return ij.FormattedLocation
	}
	
	parts := []string{}
	if ij.City != "" {
		parts = append(parts, ij.City)
	}
	if ij.State != "" {
		parts = append(parts, ij.State)
	}
	if ij.Country != "" {
		parts = append(parts, ij.Country)
	}
	
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	
	return "Sweden"
}

// detectRemote checks if job is remote
func (ic *IndeedConnector) detectRemote(ij IndeedJob) bool {
	text := strings.ToLower(ij.JobTitle + " " + ij.Snippet + " " + ij.FormattedLocation)
	
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

// parseIndeedDate parses Indeed date format
func (ic *IndeedConnector) parseIndeedDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}
	
	// Indeed uses various formats, try common ones
	formats := []string{
		time.RFC3339,
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}
	
	return time.Now()
}

// extractRequirements extracts keywords from title and snippet
func (ic *IndeedConnector) extractRequirements(ij IndeedJob) []string {
	requirements := []string{}
	seen := make(map[string]bool)
	
	text := strings.ToLower(ij.JobTitle + " " + ij.Snippet)
	
	// Common tech skills and keywords
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

// deduplicateJobs removes duplicate jobs by job key
func (ic *IndeedConnector) deduplicateJobs(jobs []models.JobPost) []models.JobPost {
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

// SyncJobs fetches jobs from Indeed and stores them
func (ic *IndeedConnector) SyncJobs() error {
	startTime := time.Now()
	fmt.Println("🔄 Starting Indeed Sweden jobs sync...")
	
	jobs, err := ic.FetchJobs()
	if err != nil {
		// Log failed sync
		ic.store.LogSync(&models.SyncLog{
			ConnectorName: ic.GetID(),
			StartedAt:     startTime,
			CompletedAt:   time.Now(),
			JobsFetched:   0,
			JobsInserted:  0,
			JobsDuplicates: 0,
			Status:        "failed",
		})
		return fmt.Errorf("failed to fetch jobs from Indeed: %w", err)
	}
	
	fmt.Printf("📥 Fetched %d jobs from Indeed Sweden\n", len(jobs))
	
	stored := 0
	duplicates := 0
	for _, job := range jobs {
		// Check if job already exists
		existing, err := ic.store.GetJob(job.ID)
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
		err = ic.store.CreateJob(&job)
		if err != nil {
			fmt.Printf("❌ Error storing job %s: %v\n", job.ID, err)
			continue
		}
		
		stored++
		fmt.Printf("✅ Stored job: %s at %s (%s)\n", job.Title, job.Company, job.Location)
	}
	
	// Log successful sync
	if err := ic.store.LogSync(&models.SyncLog{
		ConnectorName:  ic.GetID(),
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		JobsFetched:    len(jobs),
		JobsInserted:   stored,
		JobsDuplicates: duplicates,
		Status:         "success",
	}); err != nil {
		fmt.Printf("⚠️  Failed to log sync: %v\n", err)
	}
	
	fmt.Printf("🎉 Indeed sync complete! Fetched: %d, Inserted: %d, Duplicates: %d\n", len(jobs), stored, duplicates)
	return nil
}

// getLastSyncTime retrieves the timestamp of the most recent job in database
func (ic *IndeedConnector) getLastSyncTime() time.Time {
	job, err := ic.store.GetMostRecentJob("indeed-")
	if err != nil {
		fmt.Println("📅 No previous Indeed jobs found - processing all jobs")
		return time.Time{}
	}
	
	fmt.Printf("📅 Last Indeed job in database: %s (posted: %s)\n", job.Title, job.PostedDate.Format("2006-01-02"))
	return job.PostedDate
}
