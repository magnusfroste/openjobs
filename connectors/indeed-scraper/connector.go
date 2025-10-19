package indeedscraper

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"openjobs/pkg/models"
	"openjobs/pkg/storage"

	"github.com/gocolly/colly/v2"
)

// IndeedScraperConnector implements web scraping for Indeed.se
type IndeedScraperConnector struct {
	store     *storage.JobStore
	baseURL   string
	userAgent string
	rateLimit time.Duration
}

// NewIndeedScraperConnector creates a new scraper connector
func NewIndeedScraperConnector(store *storage.JobStore) *IndeedScraperConnector {
	return &IndeedScraperConnector{
		store:     store,
		baseURL:   "https://se.indeed.com",
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		rateLimit: 2 * time.Second, // Be respectful - 2 seconds between requests
	}
}

// GetID returns the connector ID
func (isc *IndeedScraperConnector) GetID() string {
	return "indeed-scraper"
}

// GetName returns the connector name
func (isc *IndeedScraperConnector) GetName() string {
	return "Indeed Sweden Scraper (Experimental)"
}

// FetchJobs scrapes job listings from Indeed.se
func (isc *IndeedScraperConnector) FetchJobs() ([]models.JobPost, error) {
	allJobs := []models.JobPost{}
	
	// Search queries for diverse coverage
	queries := []string{
		"developer",
		"engineer",
		"designer",
		"manager",
		"sales",
	}
	
	// Get last sync time for incremental sync
	lastSync := isc.getLastSyncTime()
	
	for _, query := range queries {
		fmt.Printf("🔍 Scraping Indeed for: '%s'\n", query)
		
		// Scrape first 3 pages (0, 10, 20) = 30 jobs per query
		for start := 0; start < 30; start += 10 {
			jobs, err := isc.scrapePage(query, start)
			if err != nil {
				fmt.Printf("⚠️  Error scraping page %d for query '%s': %v\n", start/10+1, query, err)
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
			
			// Rate limiting - be respectful!
			time.Sleep(isc.rateLimit)
		}
	}
	
	// Deduplicate by job key
	uniqueJobs := isc.deduplicateJobs(allJobs)
	
	fmt.Printf("📊 Scraped %d unique jobs from Indeed (filtered from %d total)\n", len(uniqueJobs), len(allJobs))
	
	return uniqueJobs, nil
}

// scrapePage scrapes a single search results page
func (isc *IndeedScraperConnector) scrapePage(query string, start int) ([]models.JobPost, error) {
	jobs := []models.JobPost{}
	
	// Build search URL
	searchURL := fmt.Sprintf("%s/jobs?q=%s&l=Sverige&start=%d",
		isc.baseURL,
		url.QueryEscape(query),
		start,
	)
	
	// Create collector with rate limiting
	c := colly.NewCollector(
		colly.UserAgent(isc.userAgent),
		colly.AllowedDomains("se.indeed.com"),
	)
	
	// Set request timeout
	c.SetRequestTimeout(30 * time.Second)
	
	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("❌ Request failed: %v\n", err)
	})
	
	// Parse job cards
	c.OnHTML("div.job_seen_beacon", func(e *colly.HTMLElement) {
		job := isc.parseJobCard(e)
		if job != nil {
			jobs = append(jobs, *job)
		}
	})
	
	// Alternative selector (Indeed changes HTML frequently)
	c.OnHTML("div[class*='jobsearch-SerpJobCard']", func(e *colly.HTMLElement) {
		job := isc.parseJobCard(e)
		if job != nil {
			jobs = append(jobs, *job)
		}
	})
	
	// Another common selector
	c.OnHTML("td.resultContent", func(e *colly.HTMLElement) {
		job := isc.parseJobCard(e)
		if job != nil {
			jobs = append(jobs, *job)
		}
	})
	
	// Visit the page
	err := c.Visit(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape page: %w", err)
	}
	
	return jobs, nil
}

// parseJobCard extracts job data from HTML element
func (isc *IndeedScraperConnector) parseJobCard(e *colly.HTMLElement) *models.JobPost {
	// Extract job title
	title := e.ChildText("h2.jobTitle span[title]")
	if title == "" {
		title = e.ChildText("h2.jobTitle")
	}
	if title == "" {
		title = e.ChildText("a[data-jk] span")
	}
	
	// Extract company name
	company := e.ChildText("span.companyName")
	if company == "" {
		company = e.ChildText("span[data-testid='company-name']")
	}
	
	// Extract location
	location := e.ChildText("div.companyLocation")
	if location == "" {
		location = e.ChildText("div[data-testid='text-location']")
	}
	
	// Extract job key (ID)
	jobKey := e.Attr("data-jk")
	if jobKey == "" {
		// Try to extract from link
		link := e.ChildAttr("a[data-jk]", "href")
		if link != "" {
			jobKey = isc.extractJobKey(link)
		}
	}
	
	// Extract snippet/description
	snippet := e.ChildText("div.job-snippet")
	if snippet == "" {
		snippet = e.ChildText("div[class*='snippet']")
	}
	if snippet == "" {
		snippet = e.ChildText("ul li")
	}
	
	// Extract salary if available
	salary := e.ChildText("div.salary-snippet")
	if salary == "" {
		salary = e.ChildText("span[class*='salary']")
	}
	
	// Skip if missing critical data
	if title == "" || jobKey == "" {
		return nil
	}
	
	// Build job URL
	jobURL := fmt.Sprintf("%s/viewjob?jk=%s", isc.baseURL, jobKey)
	
	// Create JobPost
	job := models.JobPost{
		ID:              fmt.Sprintf("indeed-scraper-%s", jobKey),
		Title:           strings.TrimSpace(title),
		Company:         strings.TrimSpace(company),
		Description:     isc.cleanSnippet(snippet),
		Location:        isc.formatLocation(location),
		Salary:          strings.TrimSpace(salary),
		SalaryMin:       nil,
		SalaryMax:       nil,
		SalaryCurrency:  "SEK",
		IsRemote:        isc.detectRemote(title, snippet, location),
		URL:             jobURL,
		EmploymentType:  "Full-time",
		ExperienceLevel: "Mid-level",
		PostedDate:      time.Now(), // Indeed doesn't show exact dates in search results
		ExpiresDate:     time.Now().AddDate(0, 1, 0),
		Requirements:    isc.extractRequirements(title, snippet),
		Benefits:        []string{},
		Fields: map[string]interface{}{
			"source":      "indeed-scraper",
			"source_url":  jobURL,
			"original_id": jobKey,
			"connector":   "indeed-scraper",
			"fetched_at":  time.Now(),
			"method":      "web_scraping",
		},
	}
	
	// Fetch full description from job page
	fullDescription := isc.scrapeJobDescription(jobURL, jobKey)
	if fullDescription != "" {
		job.Description = fullDescription
		fmt.Printf("   ✅ Fetched full description for: %s\n", title)
	}
	
	// Rate limit after fetching job page
	time.Sleep(isc.rateLimit)
	
	return &job
}

// scrapeJobDescription fetches the full job description from individual job page
func (isc *IndeedScraperConnector) scrapeJobDescription(jobURL, jobKey string) string {
	description := ""
	
	// Create a new collector for job page
	c := colly.NewCollector(
		colly.UserAgent(isc.userAgent),
		colly.AllowedDomains("se.indeed.com"),
	)
	
	c.SetRequestTimeout(30 * time.Second)
	
	// Extract full job description
	c.OnHTML("div#jobDescriptionText", func(e *colly.HTMLElement) {
		description = e.Text
	})
	
	// Alternative selector
	c.OnHTML("div.jobsearch-jobDescriptionText", func(e *colly.HTMLElement) {
		if description == "" {
			description = e.Text
		}
	})
	
	// Another common selector
	c.OnHTML("div[id*='jobDesc']", func(e *colly.HTMLElement) {
		if description == "" {
			description = e.Text
		}
	})
	
	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("   ⚠️  Failed to fetch job page %s: %v\n", jobKey, err)
	})
	
	// Visit the job page
	err := c.Visit(jobURL)
	if err != nil {
		fmt.Printf("   ⚠️  Error visiting job page %s: %v\n", jobKey, err)
		return ""
	}
	
	// Clean and return
	description = isc.cleanSnippet(description)
	return description
}

// extractJobKey extracts job key from URL
func (isc *IndeedScraperConnector) extractJobKey(link string) string {
	re := regexp.MustCompile(`jk=([a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(link)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// cleanSnippet removes HTML and cleans text
func (isc *IndeedScraperConnector) cleanSnippet(snippet string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	snippet = re.ReplaceAllString(snippet, "")
	
	// Trim whitespace
	snippet = strings.TrimSpace(snippet)
	
	return snippet
}

// formatLocation formats location string
func (isc *IndeedScraperConnector) formatLocation(location string) string {
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
func (isc *IndeedScraperConnector) detectRemote(title, snippet, location string) bool {
	text := strings.ToLower(title + " " + snippet + " " + location)
	
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

// extractRequirements extracts keywords from title and snippet
func (isc *IndeedScraperConnector) extractRequirements(title, snippet string) []string {
	requirements := []string{}
	seen := make(map[string]bool)
	
	text := strings.ToLower(title + " " + snippet)
	
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
func (isc *IndeedScraperConnector) deduplicateJobs(jobs []models.JobPost) []models.JobPost {
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

// SyncJobs scrapes jobs from Indeed and stores them
func (isc *IndeedScraperConnector) SyncJobs() error {
	startTime := time.Now()
	fmt.Println("🔄 Starting Indeed Sweden scraping sync...")
	fmt.Println("⚠️  EXPERIMENTAL: Web scraping connector")
	
	jobs, err := isc.FetchJobs()
	if err != nil {
		// Log failed sync
		isc.store.LogSync(&models.SyncLog{
			ConnectorName: isc.GetID(),
			StartedAt:     startTime,
			CompletedAt:   time.Now(),
			JobsFetched:   0,
			JobsInserted:  0,
			JobsDuplicates: 0,
			Status:        "failed",
		})
		return fmt.Errorf("failed to scrape jobs from Indeed: %w", err)
	}
	
	fmt.Printf("📥 Scraped %d jobs from Indeed Sweden\n", len(jobs))
	
	stored := 0
	duplicates := 0
	for _, job := range jobs {
		// Check if job already exists
		existing, err := isc.store.GetJob(job.ID)
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
		err = isc.store.CreateJob(&job)
		if err != nil {
			fmt.Printf("❌ Error storing job %s: %v\n", job.ID, err)
			continue
		}
		
		stored++
		fmt.Printf("✅ Stored job: %s at %s (%s)\n", job.Title, job.Company, job.Location)
	}
	
	// Log successful sync
	if err := isc.store.LogSync(&models.SyncLog{
		ConnectorName:  isc.GetID(),
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		JobsFetched:    len(jobs),
		JobsInserted:   stored,
		JobsDuplicates: duplicates,
		Status:         "success",
	}); err != nil {
		fmt.Printf("⚠️  Failed to log sync: %v\n", err)
	}
	
	fmt.Printf("🎉 Indeed scraping sync complete! Fetched: %d, Inserted: %d, Duplicates: %d\n", len(jobs), stored, duplicates)
	return nil
}

// getLastSyncTime retrieves the timestamp of the most recent job in database
func (isc *IndeedScraperConnector) getLastSyncTime() time.Time {
	job, err := isc.store.GetMostRecentJob("indeed-scraper-")
	if err != nil {
		fmt.Println("📅 No previous Indeed scraper jobs found - processing all jobs")
		return time.Time{}
	}
	
	fmt.Printf("📅 Last Indeed scraper job in database: %s (posted: %s)\n", job.Title, job.PostedDate.Format("2006-01-02"))
	return job.PostedDate
}
