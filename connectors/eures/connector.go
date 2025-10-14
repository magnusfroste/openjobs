package eures

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

// EURESConnector implements the connector interface for EURES
type EURESConnector struct {
	store     *storage.JobStore
	baseURL   string
	userAgent string
}

// EURESJob represents a job from the EURES API
type EURESJob struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Company     string `json:"company"`
	Location    struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"location"`
	Salary struct {
		Min int `json:"min,omitempty"`
		Max int `json:"max,omitempty"`
	} `json:"salary"`
	EmploymentType string    `json:"employmentType"`
	Experience     string    `json:"experience"`
	PostedDate     time.Time `json:"postedDate"`
	ExpiryDate     time.Time `json:"expiryDate"`
	URL            string    `json:"url"`
}

// EURESResponse represents the API response structure
type EURESResponse struct {
	Jobs  []EURESJob `json:"jobs"`
	Total int        `json:"total"`
}

// NewEURESConnector creates a new EURES connector
func NewEURESConnector(store *storage.JobStore) *EURESConnector {
	return &EURESConnector{
		store:     store,
		baseURL:   "https://ec.europa.eu/eures/eures-services",
		userAgent: "OpenJobs-EURES-Connector/1.0",
	}
}

// FetchJobs fetches job listings from EURES API
func (ec *EURESConnector) FetchJobs() ([]models.JobPost, error) {
	// EURES API endpoint for job search
	url := fmt.Sprintf("%s/api/jobs/search", ec.baseURL)

	// Create request with parameters
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", ec.userAgent)
	req.Header.Set("Accept", "application/json")

	// Add query parameters for recent jobs
	q := req.URL.Query()
	q.Add("sort", "date")
	q.Add("order", "desc")
	q.Add("limit", "100") // Fetch up to 100 recent jobs
	q.Add("days", "7")    // Jobs from last 7 days
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
		return nil, fmt.Errorf("EURES API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var euresResp EURESResponse
	err = json.Unmarshal(body, &euresResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Transform to our JobPost format
	jobs := make([]models.JobPost, 0, len(euresResp.Jobs))
	for _, euresJob := range euresResp.Jobs {
		job := ec.transformEURESJob(euresJob)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// transformEURESJob converts EURES job format to our JobPost format
func (ec *EURESConnector) transformEURESJob(ej EURESJob) models.JobPost {
	job := models.JobPost{
		ID:              fmt.Sprintf("eures-%s", ej.ID),
		Title:           ej.Title,
		Company:         ej.Company,
		Description:     ej.Description,
		Location:        fmt.Sprintf("%s, %s", ej.Location.City, ej.Location.Country),
		EmploymentType:  ec.mapEmploymentType(ej.EmploymentType),
		ExperienceLevel: ec.mapExperienceLevel(ej.Experience),
		PostedDate:      ej.PostedDate,
		ExpiresDate:     ej.ExpiryDate,
		Fields: map[string]interface{}{
			"source":      "eures",
			"source_url":  ej.URL,
			"original_id": ej.ID,
			"country":     ej.Location.Country,
			"city":        ej.Location.City,
			"connector":   "eures",
			"fetched_at":  time.Now(),
		},
	}

	// Handle salary
	if ej.Salary.Min > 0 || ej.Salary.Max > 0 {
		if ej.Salary.Min > 0 && ej.Salary.Max > 0 {
			job.Salary = fmt.Sprintf("€%d - €%d", ej.Salary.Min, ej.Salary.Max)
		} else if ej.Salary.Min > 0 {
			job.Salary = fmt.Sprintf("€%d+", ej.Salary.Min)
		} else {
			job.Salary = fmt.Sprintf("Up to €%d", ej.Salary.Max)
		}
	}

	return job
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
	fmt.Println("🔄 Starting EURES job sync...")

	jobs, err := ec.FetchJobs()
	if err != nil {
		return fmt.Errorf("failed to fetch jobs from EURES: %w", err)
	}

	fmt.Printf("📥 Fetched %d jobs from EURES\n", len(jobs))

	stored := 0
	for _, job := range jobs {
		// Check if job already exists (by source ID)
		existing, err := ec.store.GetJob(job.ID)
		if err != nil && err.Error() != "sql: no rows in result set" {
			fmt.Printf("⚠️  Error checking existing job %s: %v\n", job.ID, err)
			continue
		}

		if existing != nil {
			// Job already exists, skip
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

	fmt.Printf("🎉 EURES sync complete! Stored %d new jobs\n", stored)
	return nil
}
