package eures

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"openjobs/pkg/models"
	"openjobs/pkg/storage"
)

// EURESConnector implements the connector interface for EURES
type EURESConnector struct {
	store      *storage.JobStore
	baseURL    string
	userAgent  string
	appID      string
	appKey     string
	httpClient *http.Client
}

// AdzunaJob represents a job from the Adzuna API
type AdzunaJob struct {
	ID          string `json:"id"` // Changed from int to string
	Title       string `json:"title"`
	Description string `json:"description"`
	Company     struct {
		DisplayName string `json:"display_name"`
	} `json:"company"`
	Location struct {
		Area []string `json:"area"`
	} `json:"location"`
	SalaryMin    float64 `json:"salary_min,omitempty"`
	SalaryMax    float64 `json:"salary_max,omitempty"`
	ContractType string  `json:"contract_type"`
	ContractTime string  `json:"contract_time"`
	Created      string  `json:"created"`
	RedirectURL  string  `json:"redirect_url"`
}

// AdzunaResponse represents the API response structure
type AdzunaResponse struct {
	Results []AdzunaJob `json:"results"`
	Count   int         `json:"count"`
}

// GetID returns the connector ID
func (ec *EURESConnector) GetID() string {
	return "eures"
}

// GetName returns the connector name
func (ec *EURESConnector) GetName() string {
	return "EURES Connector"
}

// NewEURESConnector creates a new EURES connector
func NewEURESConnector(store *storage.JobStore) *EURESConnector {
	return &EURESConnector{
		store:      store,
		baseURL:    "https://api.adzuna.com/v1/api/jobs", // Adzuna base URL
		userAgent:  "OpenJobs-EURES-Connector/1.0",
		appID:      os.Getenv("ADZUNA_APP_ID"),
		appKey:     os.Getenv("ADZUNA_APP_KEY"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchJobs fetches job listings from Adzuna API
func (ec *EURESConnector) FetchJobs() ([]models.JobPost, error) {

	// If credentials not configured, return demo data
	if ec.appID == "" || ec.appKey == "" {
		fmt.Println("⚠️  Adzuna credentials not configured, using demo data")
		return ec.fetchDemoJobs(), nil
	}

	// Fetch from multiple European countries
	// Note: Adzuna API only supports specific countries
	// Supported: at, au, be, br, ca, ch, de, es, fr, gb, in, it, mx, nl, nz, pl, sg, us, za
	countries := []string{
		"de", // Germany
		"nl", // Netherlands
		"at", // Austria
		"ch", // Switzerland
		"be", // Belgium
		"fr", // France
		"es", // Spain
		"it", // Italy
		"pl", // Poland
		"gb", // United Kingdom
	}

	allJobs := []models.JobPost{}
	for _, country := range countries {
		countryJobs, err := ec.fetchJobsFromCountry(country)
		if err != nil {
			fmt.Printf("⚠️  Error fetching jobs from %s: %v\n", country, err)
			continue // Try next country
		}
		allJobs = append(allJobs, countryJobs...)
		fmt.Printf("   ✅ Fetched %d jobs from %s\n", len(countryJobs), country)
		
		// Rate limiting between countries
		time.Sleep(1 * time.Second)
	}

	if len(allJobs) == 0 {
		fmt.Println("⚠️  No jobs fetched from any country, using demo data")
		return ec.fetchDemoJobs(), nil
	}

	return allJobs, nil
}

// fetchJobsFromCountry fetches jobs from a specific country
func (ec *EURESConnector) fetchJobsFromCountry(country string) ([]models.JobPost, error) {
	// Get last sync time for incremental sync
	lastSync := ec.getLastSyncTime()
	
	// Build API URL with credentials - Adzuna API format
	url := fmt.Sprintf("%s/%s/search/1?app_id=%s&app_key=%s&results_per_page=100&what=developer+OR+programmer+OR+software",
		ec.baseURL, country, ec.appID, ec.appKey)
	
	// Add date filter if we have a last sync time
	if !lastSync.IsZero() {
		// Calculate days since last sync
		daysSince := int(time.Since(lastSync).Hours() / 24)
		if daysSince < 1 {
			daysSince = 1 // Minimum 1 day
		}
		url += fmt.Sprintf("&max_days_old=%d", daysSince)
		fmt.Printf("📅 Fetching jobs from last %d days\n", daysSince)
	}

	fmt.Printf("🔍 Fetching jobs from Adzuna (%s)...\n", country)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", ec.userAgent)

	resp, err := ec.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("adzuna API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var adzunaResponse AdzunaResponse
	err = json.Unmarshal(body, &adzunaResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Transform to our JobPost format
	jobs := make([]models.JobPost, 0, len(adzunaResponse.Results))
	for _, adzunaJob := range adzunaResponse.Results {
		job := ec.transformAdzunaJob(adzunaJob)
		jobs = append(jobs, job)
	}

	fmt.Printf("   ✅ Fetched %d jobs from %s\n", len(jobs), country)
	return jobs, nil
}

// fetchDemoJobs returns demo data as fallback
func (ec *EURESConnector) fetchDemoJobs() []models.JobPost {
	jobs := []models.JobPost{
		{
			ID:              "demo-001",
			Title:           "Senior Software Engineer",
			Company:         "TechCorp Sweden AB",
			Description:     "We are looking for an experienced software engineer to join our innovative team in Stockholm. You will work on cutting-edge technologies and contribute to our mission of digital transformation.",
			Location:        "Stockholm, Sweden",
			Salary:          "SEK 45,000 - 65,000/month",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			PostedDate:      time.Now().Add(-24 * time.Hour),
			ExpiresDate:     time.Now().AddDate(0, 1, 0),
			Requirements:    []string{"5+ years experience", "Go, Python, or Java", "Cloud platforms (AWS/GCP)"},
			Benefits:        []string{"Health insurance", "Flexible hours", "Remote work options"},
			Fields: map[string]interface{}{
				"source":      "demo-eures",
				"source_url":  "https://example.com/job/001",
				"original_id": "001",
				"country":     "Sweden",
				"city":        "Stockholm",
				"connector":   "eures",
				"fetched_at":  time.Now(),
			},
		},
		{
			ID:              "demo-002",
			Title:           "Full Stack Developer",
			Company:         "Nordic Startup GmbH",
			Description:     "Join our fast-growing startup as a full stack developer. We're building the future of Nordic fintech and need talented developers to help us scale.",
			Location:        "Copenhagen, Denmark",
			Salary:          "DKK 35,000 - 50,000/month",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Mid-level",
			PostedDate:      time.Now().Add(-48 * time.Hour),
			ExpiresDate:     time.Now().AddDate(0, 1, 0),
			Requirements:    []string{"3+ years experience", "React, Node.js", "PostgreSQL"},
			Benefits:        []string{"Stock options", "Learning budget", "Team events"},
			Fields: map[string]interface{}{
				"source":      "demo-eures",
				"source_url":  "https://example.com/job/002",
				"original_id": "002",
				"country":     "Denmark",
				"city":        "Copenhagen",
				"connector":   "eures",
				"fetched_at":  time.Now(),
			},
		},
		{
			ID:              "demo-003",
			Title:           "DevOps Engineer",
			Description:     "We're seeking a DevOps engineer to help us build and maintain our cloud infrastructure. Experience with Kubernetes, Docker, and CI/CD pipelines is essential.",
			Location:        "Helsinki, Finland",
			Salary:          "EUR 4,500 - 6,500/month",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			PostedDate:      time.Now().Add(-72 * time.Hour),
			ExpiresDate:     time.Now().AddDate(0, 1, 0),
			Requirements:    []string{"4+ years DevOps experience", "Kubernetes, Docker", "AWS/Azure/GCP"},
			Benefits:        []string{"Health insurance", "Professional development", "Flexible work"},
			Fields: map[string]interface{}{
				"source":      "demo-eures",
				"source_url":  "https://example.com/job/003",
				"original_id": "003",
				"country":     "Finland",
				"city":        "Helsinki",
				"connector":   "eures",
				"fetched_at":  time.Now(),
			},
		},
	}

	return jobs
}

// transformAdzunaJob converts Adzuna job format to our JobPost format
func (ec *EURESConnector) transformAdzunaJob(aj AdzunaJob) models.JobPost {
	job := models.JobPost{
		ID:              fmt.Sprintf("adzuna-%s", aj.ID),
		Title:           aj.Title,
		Company:         aj.Company.DisplayName,
		Description:     aj.Description,
		Location:        strings.Join(aj.Location.Area, ", "),
		URL:             aj.RedirectURL, // Direct application link
		EmploymentType:  ec.mapEmploymentType(aj.ContractTime),
		ExperienceLevel: "Mid-level", // Adzuna doesn't provide experience level
		PostedDate:      ec.parseAdzunaDate(aj.Created),
		ExpiresDate:     ec.parseAdzunaDate(aj.Created).AddDate(0, 1, 0), // Default 1 month expiry
		Requirements:    ec.extractRequirementsFromText(aj.Title, aj.Description), // Extract from text
		Fields: map[string]interface{}{
			"source":        "adzuna",
			"source_url":    aj.RedirectURL,
			"original_id":   aj.ID,
			"contract_type": aj.ContractType,
			"contract_time": aj.ContractTime,
			"location_area": aj.Location.Area,
			"connector":     "eures-adzuna",
			"fetched_at":    time.Now(),
		},
	}

	// Handle salary - Adzuna provides structured salary data!
	if aj.SalaryMin > 0 || aj.SalaryMax > 0 {
		// Populate structured fields for LazyJobs matching
		if aj.SalaryMin > 0 {
			salaryMin := int(aj.SalaryMin)
			job.SalaryMin = &salaryMin
		}
		if aj.SalaryMax > 0 {
			salaryMax := int(aj.SalaryMax)
			job.SalaryMax = &salaryMax
		}
		job.SalaryCurrency = "EUR"
		
		// Also create human-readable string
		if aj.SalaryMin > 0 && aj.SalaryMax > 0 {
			job.Salary = fmt.Sprintf("€%.0f - €%.0f", aj.SalaryMin, aj.SalaryMax)
		} else if aj.SalaryMin > 0 {
			job.Salary = fmt.Sprintf("€%.0f+", aj.SalaryMin)
		} else {
			job.Salary = fmt.Sprintf("Up to €%.0f", aj.SalaryMax)
		}
	}

	return job
}

// parseAdzunaDate parses Adzuna date string
func (ec *EURESConnector) parseAdzunaDate(dateStr string) time.Time {
	// Try parsing different date formats
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	// Default to current time if parsing fails
	return time.Now()
}

// mapEmploymentType converts EURES employment type to our format
func (ec *EURESConnector) mapEmploymentType(euresType string) string {
	switch strings.ToLower(euresType) {
	case "full-time", "full time":
		return "Full-time"
	case "part-time", "part time":
		return "Part-time"
	case "contract", "temporary":
		return "Contract"
	case "internship":
		return "Internship"
	default:
		return "Full-time" // Default
	}
}

// mapExperienceLevel converts EURES experience to our format
func (ec *EURESConnector) mapExperienceLevel(euresExp string) string {
	switch strings.ToLower(euresExp) {
	case "entry", "entry-level", "junior":
		return "Entry-level"
	case "mid", "mid-level", "experienced":
		return "Mid-level"
	case "senior", "expert":
		return "Senior"
	case "executive", "management":
		return "Executive"
	default:
		return "Mid-level" // Default
	}
}

// SyncJobs fetches jobs from EURES and stores them in the database
func (ec *EURESConnector) SyncJobs() error {
	startTime := time.Now()
	fmt.Println("🔄 Starting EURES job sync...")

	jobs, err := ec.FetchJobs()
	if err != nil {
		// Log failed sync
		if logErr := ec.store.LogSync(&models.SyncLog{
			ConnectorName:  ec.GetID(),
			StartedAt:      startTime,
			CompletedAt:    time.Now(),
			JobsFetched:    0,
			JobsInserted:   0,
			JobsDuplicates: 0,
			Status:         "error",
			ErrorMessage:   err.Error(),
		}); logErr != nil {
			fmt.Printf("⚠️  Failed to log sync: %v\n", logErr)
		}
		return fmt.Errorf("failed to fetch jobs from EURES: %w", err)
	}

	fmt.Printf("📥 Fetched %d jobs from EURES\n", len(jobs))

	stored := 0
	duplicates := 0
	for _, job := range jobs {
		// Check if job already exists (by source ID)
		existing, err := ec.store.GetJob(job.ID)
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
		err = ec.store.CreateJob(&job)
		if err != nil {
			fmt.Printf("❌ Error storing job %s: %v\n", job.ID, err)
			continue
		}

		stored++
		fmt.Printf("✅ Stored job: %s\n", job.Title)
	}

	// Log successful sync
	if err := ec.store.LogSync(&models.SyncLog{
		ConnectorName:  ec.GetID(),
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		JobsFetched:    len(jobs),
		JobsInserted:   stored,
		JobsDuplicates: duplicates,
		Status:         "success",
	}); err != nil {
		fmt.Printf("⚠️  Failed to log sync: %v\n", err)
	}

	fmt.Printf("🎉 EURES sync complete! Fetched: %d, Inserted: %d, Duplicates: %d\n", len(jobs), stored, duplicates)
	return nil
}

// extractRequirementsFromText extracts basic keywords from title and description
// This is a simple keyword-based extraction for APIs that don't provide structured requirements
func (ec *EURESConnector) extractRequirementsFromText(title, description string) []string {
	requirements := []string{}
	seen := make(map[string]bool)
	
	// Combine title and description for searching
	text := strings.ToLower(title + " " + description)
	
	// Common tech skills and keywords
	keywords := []string{
		// Programming languages
		"Java", "Python", "JavaScript", "TypeScript", "C++", "C#", ".NET", "PHP", "Ruby", "Go", "Rust", "Swift", "Kotlin",
		// Frameworks & Libraries
		"React", "Angular", "Vue", "Node.js", "Spring", "Django", "Flask", "Express", "Laravel",
		// DevOps & Cloud
		"Docker", "Kubernetes", "AWS", "Azure", "GCP", "CI/CD", "Jenkins", "Git", "Linux",
		// Databases
		"SQL", "PostgreSQL", "MySQL", "MongoDB", "Redis", "Elasticsearch",
		// Other skills
		"API", "REST", "GraphQL", "Microservices", "Agile", "Scrum",
	}
	
	// Extract keywords found in text
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			if !seen[keyword] {
				requirements = append(requirements, keyword)
				seen[keyword] = true
			}
		}
	}
	
	// Extract experience level from description
	if strings.Contains(text, "senior") || strings.Contains(text, "lead") {
		if !seen["Senior level"] {
			requirements = append(requirements, "Senior level")
			seen["Senior level"] = true
		}
	} else if strings.Contains(text, "junior") || strings.Contains(text, "entry") {
		if !seen["Junior level"] {
			requirements = append(requirements, "Junior level")
			seen["Junior level"] = true
		}
	}
	
	// Extract years of experience
	if strings.Contains(text, "years experience") || strings.Contains(text, "years of experience") {
		if !seen["Experience required"] {
			requirements = append(requirements, "Experience required")
			seen["Experience required"] = true
		}
	}
	
	return requirements
}

// getLastSyncTime retrieves the timestamp of the most recent job in database
func (ec *EURESConnector) getLastSyncTime() time.Time {
	job, err := ec.store.GetMostRecentJob("adzuna-")
	if err != nil {
		fmt.Println("📅 No previous EURES jobs found - fetching all jobs")
		return time.Time{}
	}
	
	fmt.Printf("📅 Last EURES job in database: %s (posted: %s)\n", job.Title, job.PostedDate.Format("2006-01-02"))
	return job.PostedDate
}
