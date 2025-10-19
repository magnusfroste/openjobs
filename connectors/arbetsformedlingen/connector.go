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

// AFJob represents a job from Arbetsförmedlingen JobSearch API
type AFJob struct {
	ID          string `json:"id"`
	Headline    string `json:"headline"`
	Description struct {
		Text            string `json:"text"`
		TextFormatted   string `json:"text_formatted"`
		Requirements    string `json:"requirements"`
		Conditions      string `json:"conditions"`
	} `json:"description"`
	Employer struct {
		Name               string `json:"name"`
		Workplace          string `json:"workplace"`
		OrganizationNumber string `json:"organization_number"`
		URL                string `json:"url"`
	} `json:"employer"`
	WorkplaceAddress struct {
		Municipality string    `json:"municipality"`
		Region       string    `json:"region"`
		Country      string    `json:"country"`
		Coordinates  []float64 `json:"coordinates"`
	} `json:"workplace_address"`
	SalaryType struct {
		Label string `json:"label"`
	} `json:"salary_type"`
	SalaryDescription string `json:"salary_description"`
	EmploymentType struct {
		Label string `json:"label"`
	} `json:"employment_type"`
	Duration struct {
		Label string `json:"label"`
	} `json:"duration"`
	WorkingHoursType struct {
		Label string `json:"label"`
	} `json:"working_hours_type"`
	ScopeOfWork struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"scope_of_work"`
	Occupation struct {
		Label string `json:"label"`
	} `json:"occupation"`
	OccupationGroup struct {
		Label string `json:"label"`
	} `json:"occupation_group"`
	OccupationField struct {
		Label string `json:"label"`
	} `json:"occupation_field"`
	MustHave struct {
		Skills []struct {
			Label string `json:"label"`
		} `json:"skills"`
		Languages []struct {
			Label string `json:"label"`
		} `json:"languages"`
		WorkExperiences []struct {
			Label string `json:"label"`
		} `json:"work_experiences"`
		Education []struct {
			Label string `json:"label"`
		} `json:"education"`
		EducationLevel []struct {
			Label string `json:"label"`
		} `json:"education_level"`
	} `json:"must_have"`
	NiceToHave struct {
		Skills []struct {
			Label string `json:"label"`
		} `json:"skills"`
		Languages []struct {
			Label string `json:"label"`
		} `json:"languages"`
		WorkExperiences []struct {
			Label string `json:"label"`
		} `json:"work_experiences"`
		Education []struct {
			Label string `json:"label"`
		} `json:"education"`
		EducationLevel []struct {
			Label string `json:"label"`
		} `json:"education_level"`
	} `json:"nice_to_have"`
	ExperienceRequired     bool `json:"experience_required"`
	DrivingLicenseRequired bool `json:"driving_license_required"`
	DrivingLicense         []struct {
		Label string `json:"label"`
	} `json:"driving_license"`
	ApplicationDetails struct {
		URL       string `json:"url"`
		Email     string `json:"email"`
		Reference string `json:"reference"`
	} `json:"application_details"`
	ApplicationDeadline string `json:"application_deadline"`
	PublicationDate     string `json:"publication_date"`
	LastPublicationDate string `json:"last_publication_date"`
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
		baseURL:   "https://jobsearch.api.jobtechdev.se",
		userAgent: "OpenJobs-Arbetsformedlingen-Connector/1.0",
	}
}

// FetchJobs fetches job listings from Arbetsförmedlingen
func (ac *ArbetsformedlingenConnector) FetchJobs() ([]models.JobPost, error) {
	// Arbetsförmedlingen JobSearch API endpoint (full data with skills)
	url := fmt.Sprintf("%s/search", ac.baseURL)

	// Create request with parameters for recent IT jobs
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", ac.userAgent)
	req.Header.Set("Accept", "application/json")

	// Get last sync time for incremental sync
	lastSync := ac.getLastSyncTime()
	
	// Add query parameters
	q := req.URL.Query()
	q.Add("q", "utvecklare OR programmer OR software") // Search for developer/programmer jobs
	q.Add("limit", "500")                             // ⭐ Increased from 20 to 500
	q.Add("sort", "pubdate-desc")                      // Sort by publication date descending
	
	// ⭐ Add timestamp filter for incremental sync
	if !lastSync.IsZero() {
		// Format: 2025-10-19 (Arbetsförmedlingen API uses this format)
		publishedAfter := lastSync.Format("2006-01-02")
		q.Add("published-after", publishedAfter)
		fmt.Printf("📅 Fetching jobs published after: %s\n", publishedAfter)
	}
	
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
		EmploymentType:  ac.mapEmploymentType(af.EmploymentType.Label),
		ExperienceLevel: ac.mapExperienceLevel(af.ExperienceRequired),
		PostedDate:      ac.parseAFDate(af.PublicationDate),
		ExpiresDate:     ac.parseAFDate(af.ApplicationDeadline),
		Requirements:    ac.extractRequirements(af),
		Benefits:        ac.extractBenefits(af),
		Fields: map[string]interface{}{
			"source":                     "arbetsformedlingen",
			"source_url":                 url,
			"original_id":                af.ID,
			"connector":                  "arbetsformedlingen",
			"language":                   ac.detectLanguage(af.Description.Text), // Detect Swedish vs English
			"fetched_at":                 time.Now(),
			// Location details
			"country":                    af.WorkplaceAddress.Country,
			"region":                     af.WorkplaceAddress.Region,
			"municipality":               af.WorkplaceAddress.Municipality,
			"coordinates":                af.WorkplaceAddress.Coordinates,
			// Occupation hierarchy
			"occupation":                 af.Occupation.Label,
			"occupation_group":           af.OccupationGroup.Label,
			"occupation_field":           af.OccupationField.Label,
			// Employment details
			"salary_type":                af.SalaryType.Label,
			"duration":                   af.Duration.Label,
			"working_hours":              af.WorkingHoursType.Label,
			"scope_of_work_min":          af.ScopeOfWork.Min,
			"scope_of_work_max":          af.ScopeOfWork.Max,
			// Application details
			"application_deadline":       af.ApplicationDeadline,
			"application_email":          af.ApplicationDetails.Email,
			"application_reference":      af.ApplicationDetails.Reference,
			"last_publication_date":      af.LastPublicationDate,
			// Requirements (structured)
			"must_have_skills":           ac.extractSkillLabels(af.MustHave.Skills),
			"nice_to_have_skills":        ac.extractSkillLabels(af.NiceToHave.Skills),
			"must_have_languages":        ac.extractSkillLabels(af.MustHave.Languages),
			"must_have_work_experiences": ac.extractSkillLabels(af.MustHave.WorkExperiences),
			"must_have_education":        ac.extractSkillLabels(af.MustHave.Education),
			"must_have_education_level":  ac.extractSkillLabels(af.MustHave.EducationLevel),
			"nice_to_have_languages":     ac.extractSkillLabels(af.NiceToHave.Languages),
			"nice_to_have_work_experiences": ac.extractSkillLabels(af.NiceToHave.WorkExperiences),
			"nice_to_have_education":     ac.extractSkillLabels(af.NiceToHave.Education),
			// Flags
			"experience_required":        af.ExperienceRequired,
			"driving_license_required":   af.DrivingLicenseRequired,
			"driving_license_types":      ac.extractSkillLabels(af.DrivingLicense),
			// Employer details
			"employer_workplace":         af.Employer.Workplace,
			"employer_organization_number": af.Employer.OrganizationNumber,
			"employer_url":               af.Employer.URL,
		},
	}

	return job
}

// extractDescription extracts job description from various fields
func (ac *ArbetsformedlingenConnector) extractDescription(af AFJob) string {
	descriptions := []string{}

	// Main description (use formatted if available, otherwise plain text)
	if af.Description.TextFormatted != "" {
		descriptions = append(descriptions, af.Description.TextFormatted)
	} else if af.Description.Text != "" {
		descriptions = append(descriptions, af.Description.Text)
	}

	// Add requirements section if available
	if af.Description.Requirements != "" {
		descriptions = append(descriptions, "Requirements: "+af.Description.Requirements)
	}

	// Add conditions section if available
	if af.Description.Conditions != "" {
		descriptions = append(descriptions, "Conditions: "+af.Description.Conditions)
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

// detectLanguage determines if job posting is in Swedish or English
// by analyzing common words in the description text
func (ac *ArbetsformedlingenConnector) detectLanguage(text string) string {
	text = strings.ToLower(text)
	
	// Swedish indicators (common words)
	swedishWords := []string{
		" är ", " och ", " för ", " med ", " att ", " du ", " vi ", " som ",
		" vill ", " söker ", " arbeta ", " företag ", " hos ", " till ", " på ",
	}
	
	// English indicators (common words)
	englishWords := []string{
		" is ", " and ", " for ", " with ", " to ", " you ", " we ", " as ",
		" want ", " looking ", " work ", " company ", " at ", " the ", " in ",
	}
	
	swedishCount := 0
	englishCount := 0
	
	for _, word := range swedishWords {
		if strings.Contains(text, word) {
			swedishCount++
		}
	}
	
	for _, word := range englishWords {
		if strings.Contains(text, word) {
			englishCount++
		}
	}
	
	// Require at least 3 matches to be confident
	if swedishCount >= 3 && swedishCount > englishCount {
		return "sv"
	}
	if englishCount >= 3 && englishCount > swedishCount {
		return "en"
	}
	
	// Default to Swedish (Arbetsförmedlingen is Swedish service)
	return "sv"
}

// extractRequirements extracts ALL job requirements for matching
// This is the SOURCE OF TRUTH - all consuming apps use this array directly
// Strategy: Extract everything that can be matched against CV skills
func (ac *ArbetsformedlingenConnector) extractRequirements(af AFJob) []string {
	requirements := []string{}

	// 1. TECHNICAL SKILLS (must-have)
	for _, skill := range af.MustHave.Skills {
		if skill.Label != "" {
			requirements = append(requirements, skill.Label)
		}
	}

	// 2. TECHNICAL SKILLS (nice-to-have)
	for _, skill := range af.NiceToHave.Skills {
		if skill.Label != "" {
			requirements = append(requirements, skill.Label)
		}
	}

	// 3. LANGUAGES (must-have) - Critical for matching!
	for _, lang := range af.MustHave.Languages {
		if lang.Label != "" {
			requirements = append(requirements, lang.Label)
		}
	}

	// 4. LANGUAGES (nice-to-have)
	for _, lang := range af.NiceToHave.Languages {
		if lang.Label != "" {
			requirements = append(requirements, lang.Label)
		}
	}

	// 5. WORK EXPERIENCES (must-have)
	for _, exp := range af.MustHave.WorkExperiences {
		if exp.Label != "" {
			requirements = append(requirements, exp.Label)
		}
	}

	// 6. WORK EXPERIENCES (nice-to-have)
	for _, exp := range af.NiceToHave.WorkExperiences {
		if exp.Label != "" {
			requirements = append(requirements, exp.Label)
		}
	}

	// 7. EDUCATION (must-have)
	for _, edu := range af.MustHave.Education {
		if edu.Label != "" {
			requirements = append(requirements, edu.Label)
		}
	}

	// 8. EDUCATION LEVEL (must-have)
	for _, level := range af.MustHave.EducationLevel {
		if level.Label != "" {
			requirements = append(requirements, level.Label)
		}
	}

	// 9. EDUCATION (nice-to-have)
	for _, edu := range af.NiceToHave.Education {
		if edu.Label != "" {
			requirements = append(requirements, edu.Label)
		}
	}

	// 10. EDUCATION LEVEL (nice-to-have)
	for _, level := range af.NiceToHave.EducationLevel {
		if level.Label != "" {
			requirements = append(requirements, level.Label)
		}
	}

	// 11. DRIVING LICENSE (if required)
	if af.DrivingLicenseRequired {
		for _, license := range af.DrivingLicense {
			if license.Label != "" {
				requirements = append(requirements, license.Label)
			}
		}
	}

	// 12. OCCUPATION (always add for categorization)
	if af.Occupation.Label != "" {
		requirements = append(requirements, af.Occupation.Label)
	}

	// 13. OCCUPATION GROUP (for broader matching)
	if af.OccupationGroup.Label != "" {
		requirements = append(requirements, af.OccupationGroup.Label)
	}

	// 14. EXPERIENCE FLAG
	if af.ExperienceRequired {
		requirements = append(requirements, "Work experience required")
	}

	// 15. STRUCTURED REQUIREMENTS from description
	if af.Description.Requirements != "" {
		requirements = append(requirements, af.Description.Requirements)
	}

	return requirements
}

// extractBenefits extracts job benefits
func (ac *ArbetsformedlingenConnector) extractBenefits(af AFJob) []string {
	benefits := []string{}

	// Add employment type as benefit if permanent
	if strings.Contains(strings.ToLower(af.EmploymentType.Label), "tillsvidare") ||
		strings.Contains(strings.ToLower(af.EmploymentType.Label), "permanent") {
		benefits = append(benefits, "Permanent employment")
	}

	// Add duration info
	if af.Duration.Label != "" {
		benefits = append(benefits, af.Duration.Label)
	}

	// Add working hours type
	if af.WorkingHoursType.Label != "" {
		benefits = append(benefits, af.WorkingHoursType.Label)
	}

	return benefits
}

// extractURL extracts the job URL
func (ac *ArbetsformedlingenConnector) extractURL(af AFJob) string {
	// Use application URL if available
	if af.ApplicationDetails.URL != "" {
		return af.ApplicationDetails.URL
	}
	// Fallback to Arbetsförmedlingen page
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
	startTime := time.Now()
	fmt.Println("🔄 Starting Arbetsförmedlingen job sync...")

	jobs, err := ac.FetchJobs()
	if err != nil {
		// Log failed sync
		ac.store.LogSync(&models.SyncLog{
			ConnectorName: ac.GetID(),
			StartedAt:     startTime,
			CompletedAt:   time.Now(),
			JobsFetched:   0,
			JobsInserted:  0,
			JobsDuplicates: 0,
			Status:        "error",
			ErrorMessage:  err.Error(),
		})
		return fmt.Errorf("failed to fetch jobs from Arbetsförmedlingen: %w", err)
	}

	fmt.Printf("📥 Fetched %d jobs from Arbetsförmedlingen\n", len(jobs))

	stored := 0
	duplicates := 0
	for _, job := range jobs {
		// Check if job already exists
		existing, err := ac.store.GetJob(job.ID)
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
		err = ac.store.CreateJob(&job)
		if err != nil {
			fmt.Printf("❌ Error storing job %s: %v\n", job.ID, err)
			continue
		}

		stored++
		fmt.Printf("✅ Stored job: %s at %s\n", job.Title, job.Company)
	}

	// Log successful sync
	if err := ac.store.LogSync(&models.SyncLog{
		ConnectorName:  ac.GetID(),
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		JobsFetched:    len(jobs),
		JobsInserted:   stored,
		JobsDuplicates: duplicates,
		Status:         "success",
	}); err != nil {
		fmt.Printf("⚠️  Failed to log sync: %v\n", err)
	}

	// ⭐ Save sync timestamp for next incremental sync
	if err := ac.saveLastSyncTime(); err != nil {
		fmt.Printf("⚠️  Failed to save sync timestamp: %v\n", err)
	}
	
	fmt.Printf("🎉 Arbetsförmedlingen sync complete! Fetched: %d, Inserted: %d, Duplicates: %d\n", len(jobs), stored, duplicates)
	return nil
}

// extractSkillLabels is a helper to extract labels from skill structs
func (ac *ArbetsformedlingenConnector) extractSkillLabels(skills interface{}) []string {
	labels := []string{}
	
	// Use reflection to handle different struct types
	switch v := skills.(type) {
	case []struct{ Label string `json:"label"` }:
		for _, skill := range v {
			if skill.Label != "" {
				labels = append(labels, skill.Label)
			}
		}
	}
	
	return labels
}

// getLastSyncTime retrieves the timestamp of the most recent job in database
// This is used for incremental sync - only fetch jobs newer than this
func (ac *ArbetsformedlingenConnector) getLastSyncTime() time.Time {
	// Query the most recent job's posted_date from our connector
	job, err := ac.store.GetMostRecentJob("af-")
	if err != nil {
		// No jobs yet or error - this is first sync
		fmt.Println("📅 No previous jobs found - fetching all jobs")
		return time.Time{}
	}
	
	fmt.Printf("📅 Last job in database: %s (posted: %s)\n", job.Title, job.PostedDate.Format("2006-01-02"))
	return job.PostedDate
}

// saveLastSyncTime is no longer needed - database tracks this automatically
// Keeping empty function for compatibility
func (ac *ArbetsformedlingenConnector) saveLastSyncTime() error {
	// Database automatically tracks via posted_date - no action needed
	return nil
}
