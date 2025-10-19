package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	// Test scraping the specific job you mentioned
	jobURL := "https://se.indeed.com/viewjob?jk=06a33c7f0be7fcaf"
	
	fmt.Println("🔍 Testing full job description scraping...")
	fmt.Printf("📄 Job URL: %s\n\n", jobURL)
	
	description := scrapeJobDescription(jobURL)
	
	if description != "" {
		fmt.Println("✅ SUCCESS! Full description extracted:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println(description)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("\n📊 Description length: %d characters\n", len(description))
		fmt.Printf("📊 Word count: ~%d words\n", len(strings.Fields(description)))
	} else {
		fmt.Println("❌ Failed to extract description")
	}
}

func scrapeJobDescription(jobURL string) string {
	description := ""
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	
	// Create a new collector
	c := colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.AllowedDomains("se.indeed.com"),
	)
	
	c.SetRequestTimeout(30 * time.Second)
	
	// Extract full job description - try multiple selectors
	c.OnHTML("div#jobDescriptionText", func(e *colly.HTMLElement) {
		description = e.Text
		fmt.Println("✅ Found description with selector: div#jobDescriptionText")
	})
	
	c.OnHTML("div.jobsearch-jobDescriptionText", func(e *colly.HTMLElement) {
		if description == "" {
			description = e.Text
			fmt.Println("✅ Found description with selector: div.jobsearch-jobDescriptionText")
		}
	})
	
	c.OnHTML("div[id*='jobDesc']", func(e *colly.HTMLElement) {
		if description == "" {
			description = e.Text
			fmt.Println("✅ Found description with selector: div[id*='jobDesc']")
		}
	})
	
	// Try broader selector
	c.OnHTML("div[class*='jobsearch']", func(e *colly.HTMLElement) {
		if description == "" && len(e.Text) > 200 {
			description = e.Text
			fmt.Println("✅ Found description with selector: div[class*='jobsearch']")
		}
	})
	
	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("❌ Request failed: %v\n", err)
		fmt.Printf("   Status code: %d\n", r.StatusCode)
	})
	
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("🌐 Visiting: %s\n", r.URL.String())
	})
	
	// Visit the job page
	err := c.Visit(jobURL)
	if err != nil {
		fmt.Printf("❌ Error visiting job page: %v\n", err)
		return ""
	}
	
	// Clean the description
	description = cleanSnippet(description)
	return description
}

func cleanSnippet(snippet string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	snippet = re.ReplaceAllString(snippet, "")
	
	// Trim whitespace
	snippet = strings.TrimSpace(snippet)
	
	return snippet
}
