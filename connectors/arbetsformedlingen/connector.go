package arbetsformedlingen

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"openjobs/pkg/models"
	"openjobs/pkg/storage"
)

// ArbetsformedlingenConnector implements connector for Swedish employment service
type ArbetsformedlingenConnector struct {
	store     *storage.JobStore
	baseURL   string
	userAgent string
}

// AFJob represents a job from Arbetsförmedlingen API
type AFJob struct {
	ID          string `json:"id"`
	Headline    string `json:"headline"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
	Employer struct {
		Name string `json:"name"`
	} `json:"employer"`
	WorkplaceAddress struct {
		Municipality string `json:"municipality"`
		Region       string `json:"region"`
		Country      string `json:"country"`
	} `json:"workplace_address"`
	SalaryDescription string `json:"salary_description"`
	EmploymentType    struct {
		ConceptLabel string `json:"concept_label"`
	} `json:"employment_type"`
	ExperienceRequired bool `json:"experience_required"`
	ApplicationDetails struct {
		Information []struct {
			Headline string `json:"headline"`
			Text     string `json:"text"`
		} `json:"information"`
	} `json:"application_details"`
	PublicationDate     string `json:"publication_date"`
	LastApplicationDate string `json:"last_application_date"`
	SourceLinks         []struct {
		URL string `json:"url"`
	} `json:"source_links"`
}

// AFResponse represents the API response
type AFResponse struct {
	Total struct {
		Value int `json:"value"`
	} `json:"total"`
	Hits []AFJob `json:"hits"`
}

// GetID returns the connector ID
func (ac *ArbetsformedlingenConnector) GetID() string {
	return "arbetsformedlingen"
}

// GetName returns the connector name
func (ac *ArbetsformedlingenConnector) GetName() string {
	return "Arbetsförmedlingen Connector"
}

// NewArbetsformedlingenConnector creates a new connector
func NewArbetsformedlingenConnector(store *storage.JobStore) *ArbetsformedlingenConnector {
	return &ArbetsformedlingenConnector{
		store:     store,
		baseURL:   "https://links.api.jobtechdev.se",
		userAgent: "OpenJobs-Arbetsformedlingen-Connector/1.0",
	}
}

// FetchJobs fetches job listings from Arbetsförmedlingen
func (ac *ArbetsformedlingenConnector) FetchJobs() ([]models.JobPost, error) {
	// Arbetsförmedlingen JobTech API endpoint
	url := fmt.Sprintf("%s/joblinks", ac.baseURL)

	// Create request with parameters for recent IT jobs
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", ac.userAgent)
	req.Header.Set("Accept", "application/json")

	// Add query parameters
	q := req.URL.Query()
	q.Add("q", "utvecklare OR programmer OR software") // Search for developer/programmer jobs
	q.Add("limit", "20")                               // Get 20 jobs
	q.Add("sort", "pubdate-desc")                      // Sort by publication date descending
	req.URL.RawQuery = q.Encode()

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Arbetsförmedlingen API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var afResponse AFResponse
	err = json.Unmarshal(body, &afResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Transform to our JobPost format
	jobs := make([]models.JobPost, 0, len(afResponse.Hits))
	for _, afJob := range afResponse.Hits {
		job := ac.transformAFJob(afJob)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// transformAFJob converts Arbetsförmedlingen job format to our JobPost format
func (ac *ArbetsformedlingenConnector) transformAFJob(af AFJob) models.JobPost {
	// Parse salary
	salaryMin, salaryMax, currency := ac.parseSalary(af.SalaryDescription)
	
	// Detect remote work
	isRemote := ac.detectRemote(af)
	
	// Extract URL
	url := ac.extractURL(af)
	
	job := models.JobPost{
		ID:              fmt.Sprintf("af-%s", af.ID),
		Title:           af.Headline,
		Company:         af.Employer.Name,
		Description:     ac.extractDescription(af),
		Location:        ac.formatLocation(af),
		Salary:          af.SalaryDescription,
		SalaryMin:       salaryMin,
		SalaryMax:       salaryMax,
		SalaryCurrency:  currency,
		IsRemote:        isRemote,
		URL:             url,
		EmploymentType:  ac.mapEmploymentType(af.EmploymentType.ConceptLabel),
		ExperienceLevel: ac.mapExperienceLevel(af.ExperienceRequired),
		PostedDate:      ac.parseAFDate(af.PublicationDate),
		ExpiresDate:     ac.parseAFDate(af.LastApplicationDate),
		Requirements:    ac.extractRequirements(af),
		Benefits:        ac.extractBenefits(af),
		Fields: map[string]interface{}{
			"source":       "arbetsformedlingen",
			"source_url":   url,
			"original_id":  af.ID,
			"country":      af.WorkplaceAddress.Country,
			"region":       af.WorkplaceAddress.Region,
			"municipality": af.WorkplaceAddress.Municipality,
			"connector":    "arbetsformedlingen",
			"fetched_at":   time.Now(),
		},
	}

	return job
}

// extractDescription extracts job description from various fields
func (ac *ArbetsformedlingenConnector) extractDescription(af AFJob) string {
	descriptions := []string{}

	// Main description
	if af.Description.Text != "" {
		descriptions = append(descriptions, af.Description.Text)
	}

	// Additional information from application details
	for _, info := range af.ApplicationDetails.Information {
		if info.Text != "" {
			descriptions = append(descriptions, fmt.Sprintf("%s: %s", info.Headline, info.Text))
		}
	}

	return strings.Join(descriptions, "\n\n")
}

// formatLocation formats the job location
func (ac *ArbetsformedlingenConnector) formatLocation(af AFJob) string {
	parts := []string{}
	if af.WorkplaceAddress.Municipality != "" {
		parts = append(parts, af.WorkplaceAddress.Municipality)
	}
	if af.WorkplaceAddress.Region != "" {
		parts = append(parts, af.WorkplaceAddress.Region)
	}
	if af.WorkplaceAddress.Country != "" {
		parts = append(parts, af.WorkplaceAddress.Country)
	}
	return strings.Join(parts, ", ")
}

// mapEmploymentType converts AF employment type to our format
func (ac *ArbetsformedlingenConnector) mapEmploymentType(afType string) string {
	switch strings.ToLower(afType) {
	case "heltid":
		return "Full-time"
	case "deltid":
		return "Part-time"
	case "vikariat", "projekt":
		return "Contract"
	default:
		return "Full-time" // Default
	}
}

// mapExperienceLevel converts AF experience to our format
func (ac *ArbetsformedlingenConnector) mapExperienceLevel(required bool) string {
	if required {
		return "Mid-level"
	}
	return "Entry-level"
}

// extractRequirements extracts job requirements
func (ac *ArbetsformedlingenConnector) extractRequirements(af AFJob) []string {
	requirements := []string{}

	if af.ExperienceRequired {
		requirements = append(requirements, "Work experience required")
	}

	// Look for requirements in application details
	for _, info := range af.ApplicationDetails.Information {
		if strings.Contains(strings.ToLower(info.Headline), "krav") ||
			strings.Contains(strings.ToLower(info.Headline), "requirements") {
			requirements = append(requirements, info.Text)
		}
	}

	return requirements
}

// extractBenefits extracts job benefits
func (ac *ArbetsformedlingenConnector) extractBenefits(af AFJob) []string {
	benefits := []string{}

	// Look for benefits in application details
	for _, info := range af.ApplicationDetails.Information {
		if strings.Contains(strings.ToLower(info.Headline), "förmån") ||
			strings.Contains(strings.ToLower(info.Headline), "benefit") {
			benefits = append(benefits, info.Text)
		}
	}

	return benefits
}

// extractURL extracts the job URL
func (ac *ArbetsformedlingenConnector) extractURL(af AFJob) string {
	if len(af.SourceLinks) > 0 {
		return af.SourceLinks[0].URL
	}
	return fmt.Sprintf("https://arbetsformedlingen.se/platsbanken/annonser/%s", af.ID)
}

// parseSalary parses salary string to extract min, max, and currency
func (ac *ArbetsformedlingenConnector) parseSalary(salaryStr string) (*int, *int, string) {
	if salaryStr == "" {
		return nil, nil, "SEK"
	}

	// Common patterns in Swedish job postings:
	// "45000 - 65000 SEK/månad"
	// "SEK 45,000 - 65,000"
	// "45 000 - 65 000 kr/mån"
	// "Enligt överenskommelse" (by agreement)

	// Check if it's "by agreement" or similar
	lower := strings.ToLower(salaryStr)
	if strings.Contains(lower, "överenskommelse") || 
	   strings.Contains(lower, "agreement") ||
	   strings.Contains(lower, "enligt ök") {
		return nil, nil, "SEK"
	}

	// Remove common Swedish/English words
	cleanStr := strings.ReplaceAll(salaryStr, "kr", "")
	cleanStr = strings.ReplaceAll(cleanStr, "SEK", "")
	cleanStr = strings.ReplaceAll(cleanStr, "månad", "")
	cleanStr = strings.ReplaceAll(cleanStr, "mån", "")
	cleanStr = strings.ReplaceAll(cleanStr, "month", "")
	cleanStr = strings.ReplaceAll(cleanStr, "/", "")
	cleanStr = strings.ReplaceAll(cleanStr, ",", "")
	cleanStr = strings.ReplaceAll(cleanStr, " ", "")

	// Try to find two numbers separated by dash or "till"
	var min, max int
	var currency string = "SEK"

	// Pattern: "45000-65000" or "45000till65000"
	if strings.Contains(cleanStr, "-") {
		parts := strings.Split(cleanStr, "-")
		if len(parts) == 2 {
			if m, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
				min = m
			}
			if m, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				max = m
			}
		}
	} else if strings.Contains(strings.ToLower(cleanStr), "till") {
		parts := strings.Split(strings.ToLower(cleanStr), "till")
		if len(parts) == 2 {
			if m, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
				min = m
			}
			if m, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				max = m
			}
		}
	} else {
		// Single number - use as both min and max
		if m, err := strconv.Atoi(cleanStr); err == nil {
			min = m
			max = m
		}
	}

	if min > 0 && max > 0 {
		return &min, &max, currency
	} else if min > 0 {
		return &min, nil, currency
	}

	return nil, nil, currency
}

// detectRemote checks if the job allows remote work
func (ac *ArbetsformedlingenConnector) detectRemote(af AFJob) bool {
	// Check location
	location := strings.ToLower(ac.formatLocation(af))
	description := strings.ToLower(af.Description.Text)
	headline := strings.ToLower(af.Headline)

	// Swedish and English keywords for remote work
	remoteKeywords := []string{
		"distans", "remote", "hemarbete", "hemifrån",
		"fjärr", "work from home", "wfh",
		"anywhere", "var som helst",
	}

	for _, keyword := range remoteKeywords {
		if strings.Contains(location, keyword) ||
		   strings.Contains(description, keyword) ||
		   strings.Contains(headline, keyword) {
			return true
		}
	}

	return false
}

// parseAFDate parses Arbetsförmedlingen date string
func (ac *ArbetsformedlingenConnector) parseAFDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}

	// Try different date formats
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

	return time.Now()
}

// SyncJobs fetches jobs from Arbetsförmedlingen and stores them
func (ac *ArbetsformedlingenConnector) SyncJobs() error {
	fmt.Println("🔄 Starting Arbetsförmedlingen job sync...")

	jobs, err := ac.FetchJobs()
	if err != nil {
		return fmt.Errorf("failed to fetch jobs from Arbetsförmedlingen: %w", err)
	}

	fmt.Printf("📥 Fetched %d jobs from Arbetsförmedlingen\n", len(jobs))

	stored := 0
	for _, job := range jobs {
		// Check if job already exists
		existing, err := ac.store.GetJob(job.ID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			fmt.Printf("⚠️  Error checking existing job %s: %v\n", job.ID, err)
			continue
		}

		if existing != nil {
			// Job already exists, skip
			continue
		}

		// Store new job
		err = ac.store.CreateJob(&job)
		if err != nil {
			fmt.Printf("❌ Error storing job %s: %v\n", job.ID, err)
			continue
		}

		stored++
		fmt.Printf("✅ Stored job: %s at %s\n", job.Title, job.Company)
	}

	fmt.Printf("🎉 Arbetsförmedlingen sync complete! Stored %d new jobs\n", stored)
	return nil
}
