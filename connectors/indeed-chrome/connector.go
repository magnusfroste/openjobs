package indeedchrome

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"openjobs/pkg/models"
	"openjobs/pkg/storage"

	"github.com/chromedp/chromedp"
)

// IndeedChromeConnector implements headless Chrome scraping for Indeed.se
type IndeedChromeConnector struct {
	store     *storage.JobStore
	baseURL   string
	rateLimit time.Duration
}

// NewIndeedChromeConnector creates a new Chrome-based scraper connector
func NewIndeedChromeConnector(store *storage.JobStore) *IndeedChromeConnector {
	return &IndeedChromeConnector{
		store:     store,
		baseURL:   "https://se.indeed.com",
		rateLimit: 3 * time.Second, // Be extra respectful with Chrome
	}
}

// GetID returns the connector ID
func (icc *IndeedChromeConnector) GetID() string {
	return "indeed-chrome"
}

// GetName returns the connector name
func (icc *IndeedChromeConnector) GetName() string {
	return "Indeed Sweden Chrome Scraper (Headless Browser)"
}

// FetchJobs scrapes job listings from Indeed.se using headless Chrome
func (icc *IndeedChromeConnector) FetchJobs() ([]models.JobPost, error) {
	allJobs := []models.JobPost{}
	
	// Search queries for diverse coverage
	queries := []string{
		"developer",
		"engineer",
		"designer",
		"manager",
		"sales",
		"marketing",
	}
	
	// Get existing job IDs for incremental sync (database-based)
	existingIDs := icc.getExistingJobIDs()
	
	// Create shared Chrome context for all pages (reuse to save memory)
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("headless", true),
		chromedp.ExecPath("/usr/bin/chromium-browser"),
	}
	
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()
	
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()
	
	for _, query := range queries {
		fmt.Printf("🔍 Scraping Indeed with Chrome for: '%s'\n", query)
		
		duplicateCount := 0
		maxDuplicatesBeforeStop := 20 // Stop if we see 20 duplicates in a row
		
		// Scrape first 10 pages (0-90) = 100 jobs per query
		// Since we run once per day, maximize coverage
		for start := 0; start < 100; start += 10 {
			jobs, err := icc.scrapePage(browserCtx, query, start)
			if err != nil {
				fmt.Printf("⚠️  Error scraping page %d for query '%s': %v\n", start/10+1, query, err)
				continue
			}
			
			if len(jobs) == 0 {
				fmt.Printf("   ℹ️  No more results for '%s' at page %d\n", query, start/10+1)
				break // No more results
			}
			
			newJobsOnPage := 0
			// Filter to only new jobs (check against existing IDs)
			for _, job := range jobs {
				if !existingIDs[job.ID] {
					// This is a new job - fetch full description
					fullJob := icc.createJobPost(browserCtx, map[string]string{
						"jobKey":  strings.TrimPrefix(job.ID, "indeed-chrome-"),
						"title":   job.Title,
						"company": job.Company,
						"location": job.Location,
						"snippet": job.Description,
						"salary":  job.Salary,
					}, true) // true = fetch full description
					
					if fullJob != nil {
						allJobs = append(allJobs, *fullJob)
						newJobsOnPage++
					}
				}
			}
			
			// Early stopping: if we found no new jobs on this page, increment duplicate counter
			if newJobsOnPage == 0 {
				duplicateCount += len(jobs)
				fmt.Printf("   ℹ️  Page %d: All %d jobs already seen (total duplicates: %d)\n", start/10+1, len(jobs), duplicateCount)
				
				// Stop if we've seen too many duplicates (means we've caught up)
				if duplicateCount >= maxDuplicatesBeforeStop {
					fmt.Printf("   ✅ Stopping '%s' - caught up with existing jobs\n", query)
					break
				}
			} else {
				duplicateCount = 0 // Reset counter if we found new jobs
				fmt.Printf("   ✅ Page %d: Found %d new jobs\n", start/10+1, newJobsOnPage)
			}
			
			// Rate limiting - be respectful!
			time.Sleep(icc.rateLimit)
		}
	}
	
	// Deduplicate by job key
	uniqueJobs := icc.deduplicateJobs(allJobs)
	
	fmt.Printf("📊 Scraped %d unique jobs from Indeed (filtered from %d total)\n", len(uniqueJobs), len(allJobs))
	
	return uniqueJobs, nil
}

// scrapePage scrapes a single search results page using Chrome
func (icc *IndeedChromeConnector) scrapePage(browserCtx context.Context, query string, start int) ([]models.JobPost, error) {
	jobs := []models.JobPost{}
	
	// Build search URL
	searchURL := fmt.Sprintf("%s/jobs?q=%s&l=Sverige&start=%d",
		icc.baseURL,
		strings.ReplaceAll(query, " ", "+"),
		start,
	)
	
	// Create timeout context for this page
	ctx, timeoutCancel := context.WithTimeout(browserCtx, 120*time.Second)
	defer timeoutCancel()
	
	// Variables to store scraped data
	var jobCards []map[string]string
	
	// Run Chrome automation
	err := chromedp.Run(ctx,
		// Navigate to search page
		chromedp.Navigate(searchURL),
		
		// Wait for job cards to load
		chromedp.WaitVisible(`div.job_seen_beacon, td.resultContent`, chromedp.ByQuery),
		
		// Wait a bit for dynamic content
		chromedp.Sleep(2*time.Second),
		
		// Extract job data using JavaScript
		chromedp.Evaluate(`
			(() => {
				const jobs = [];
				
				// Try multiple selectors
				const cards = document.querySelectorAll('div.job_seen_beacon, td.resultContent, div[class*="jobsearch"]');
				
				cards.forEach(card => {
					// Extract job key
					let jobKey = card.getAttribute('data-jk');
					if (!jobKey) {
						const link = card.querySelector('a[data-jk]');
						if (link) jobKey = link.getAttribute('data-jk');
					}
					if (!jobKey) {
						const href = card.querySelector('a[href*="jk="]');
						if (href) {
							const match = href.href.match(/jk=([a-zA-Z0-9]+)/);
							if (match) jobKey = match[1];
						}
					}
					
					// Extract title
					let title = '';
					const titleEl = card.querySelector('h2.jobTitle span[title], h2.jobTitle, a[data-jk] span');
					if (titleEl) title = titleEl.textContent.trim();
					
					// Extract company
					let company = '';
					const companyEl = card.querySelector('span.companyName, span[data-testid="company-name"]');
					if (companyEl) company = companyEl.textContent.trim();
					
					// Extract location
					let location = '';
					const locationEl = card.querySelector('div.companyLocation, div[data-testid="text-location"]');
					if (locationEl) location = locationEl.textContent.trim();
					
					// Extract snippet
					let snippet = '';
					const snippetEl = card.querySelector('div.job-snippet, div[class*="snippet"], ul li');
					if (snippetEl) snippet = snippetEl.textContent.trim();
					
					// Extract salary
					let salary = '';
					const salaryEl = card.querySelector('div.salary-snippet, span[class*="salary"]');
					if (salaryEl) salary = salaryEl.textContent.trim();
					
					// Only add if we have minimum data
					if (title && jobKey) {
						jobs.push({
							jobKey: jobKey,
							title: title,
							company: company,
							location: location,
							snippet: snippet,
							salary: salary
						});
					}
				});
				
				return jobs;
			})()
		`, &jobCards),
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to scrape page with Chrome: %w", err)
	}
	
	fmt.Printf("   📄 Found %d job cards on page\n", len(jobCards))
	
	// Convert to JobPost objects (without full descriptions yet)
	for _, card := range jobCards {
		job := icc.createJobPost(browserCtx, card, false) // false = don't fetch description yet
		if job != nil {
			jobs = append(jobs, *job)
		}
	}
	
	return jobs, nil
}

// createJobPost creates a JobPost from scraped data
func (icc *IndeedChromeConnector) createJobPost(browserCtx context.Context, card map[string]string, fetchDescription bool) *models.JobPost {
	jobKey := card["jobKey"]
	title := card["title"]
	company := card["company"]
	location := card["location"]
	snippet := card["snippet"]
	salary := card["salary"]
	
	if title == "" || jobKey == "" {
		return nil
	}
	
	// Build job URL
	jobURL := fmt.Sprintf("%s/viewjob?jk=%s", icc.baseURL, jobKey)
	
	// Optionally fetch full description from job page
	description := snippet
	if fetchDescription {
		fullDescription := icc.scrapeJobDescription(browserCtx, jobURL, jobKey)
		if fullDescription != "" {
			description = fullDescription
			fmt.Printf("   ✅ Fetched full description for: %s\n", title)
		}
	}
	
	// Estimate posted date (Indeed doesn't always show exact date)
	// Use a heuristic: jobs are likely posted within last 30 days
	// For incremental sync, we'll check against database ID instead
	postedDate := time.Now().AddDate(0, 0, -7) // Assume posted within last week
	
	// Create JobPost
	job := models.JobPost{
		ID:              fmt.Sprintf("indeed-chrome-%s", jobKey),
		Title:           strings.TrimSpace(title),
		Company:         strings.TrimSpace(company),
		Description:     icc.cleanText(description),
		Location:        icc.formatLocation(location),
		Salary:          strings.TrimSpace(salary),
		SalaryMin:       nil,
		SalaryMax:       nil,
		SalaryCurrency:  "SEK",
		IsRemote:        icc.detectRemote(title, description, location),
		URL:             jobURL,
		EmploymentType:  "Full-time",
		ExperienceLevel: "Mid-level",
		PostedDate:      postedDate,
		ExpiresDate:     time.Now().AddDate(0, 1, 0),
		Requirements:    icc.extractRequirements(title, description),
		Benefits:        []string{},
		Fields: map[string]interface{}{
			"source":      "indeed-chrome",
			"source_url":  jobURL,
			"original_id": jobKey,
			"connector":   "indeed-chrome",
			"fetched_at":  time.Now(),
			"method":      "headless_chrome",
		},
	}
	
	return &job
}

// scrapeJobDescription fetches the full job description from individual job page using Chrome
func (icc *IndeedChromeConnector) scrapeJobDescription(browserCtx context.Context, jobURL, jobKey string) string {
	description := ""
	
	// Create timeout context for this job page
	ctx, timeoutCancel := context.WithTimeout(browserCtx, 60*time.Second)
	defer timeoutCancel()
	
	// Run Chrome automation
	err := chromedp.Run(ctx,
		// Navigate to job page
		chromedp.Navigate(jobURL),
		
		// Wait for description to load
		chromedp.WaitVisible(`div#jobDescriptionText, div.jobsearch-jobDescriptionText`, chromedp.ByQuery),
		
		// Extract description text
		chromedp.Evaluate(`
			(() => {
				const descEl = document.querySelector('div#jobDescriptionText, div.jobsearch-jobDescriptionText, div[id*="jobDesc"]');
				return descEl ? descEl.textContent : '';
			})()
		`, &description),
	)
	
	if err != nil {
		fmt.Printf("   ⚠️  Failed to fetch job page %s: %v\n", jobKey, err)
		return ""
	}
	
	return description
}

// cleanText removes extra whitespace and cleans text
func (icc *IndeedChromeConnector) cleanText(text string) string {
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

// formatLocation formats location string
func (icc *IndeedChromeConnector) formatLocation(location string) string {
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
func (icc *IndeedChromeConnector) detectRemote(title, description, location string) bool {
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
func (icc *IndeedChromeConnector) extractRequirements(title, description string) []string {
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
func (icc *IndeedChromeConnector) deduplicateJobs(jobs []models.JobPost) []models.JobPost {
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

// SyncJobs scrapes jobs from Indeed using Chrome and stores them
func (icc *IndeedChromeConnector) SyncJobs() error {
	startTime := time.Now()
	fmt.Println("🔄 Starting Indeed Sweden Chrome scraping sync...")
	fmt.Println("🌐 Using headless Chrome - bypasses Cloudflare!")
	
	jobs, err := icc.FetchJobs()
	if err != nil {
		// Log failed sync
		icc.store.LogSync(&models.SyncLog{
			ConnectorName: icc.GetID(),
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
		existing, err := icc.store.GetJob(job.ID)
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
		err = icc.store.CreateJob(&job)
		if err != nil {
			fmt.Printf("❌ Error storing job %s: %v\n", job.ID, err)
			continue
		}
		
		stored++
		fmt.Printf("✅ Stored job: %s at %s (%s)\n", job.Title, job.Company, job.Location)
	}
	
	// Log successful sync
	if err := icc.store.LogSync(&models.SyncLog{
		ConnectorName:  icc.GetID(),
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		JobsFetched:    len(jobs),
		JobsInserted:   stored,
		JobsDuplicates: duplicates,
		Status:         "success",
	}); err != nil {
		fmt.Printf("⚠️  Failed to log sync: %v\n", err)
	}
	
	fmt.Printf("🎉 Indeed Chrome scraping sync complete! Fetched: %d, Inserted: %d, Duplicates: %d\n", len(jobs), stored, duplicates)
	return nil
}

// getExistingJobIDs retrieves all existing job IDs for incremental sync
func (icc *IndeedChromeConnector) getExistingJobIDs() map[string]bool {
	// This would need a new method in storage to get all IDs efficiently
	// For now, use a simple approach
	existingIDs := make(map[string]bool)
	
	// Note: In production, you'd want to add a method to JobStore like:
	// jobs, err := icc.store.GetJobIDsByPrefix("indeed-chrome-")
	// For now, we'll rely on the database check in SyncJobs
	
	fmt.Println("📅 Using database-based incremental sync")
	return existingIDs
}
