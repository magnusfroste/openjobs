package models

import "time"

// AnalyticsSummary provides high-level overview statistics
type AnalyticsSummary struct {
	TotalJobs        int  `json:"total_jobs"`
	SourcesCount     int  `json:"sources_count"`
	CountriesCovered int  `json:"countries_covered"`
	AvgSalaryRange   *int `json:"avg_salary_range,omitempty"`
	RemotePercentage *int `json:"remote_percentage,omitempty"`
}

// SourceAnalytics provides detailed statistics per job source
type SourceAnalytics struct {
	Source           string     `json:"source"`
	TotalJobs        int        `json:"total_jobs"`
	CountriesCovered int        `json:"countries_covered"`
	FulltimeJobs     int        `json:"fulltime_jobs"`
	ParttimeJobs     int        `json:"parttime_jobs"`
	RemoteJobs       int        `json:"remote_jobs"`
	AvgMinSalary     *int       `json:"avg_min_salary,omitempty"`
	AvgMaxSalary     *int       `json:"avg_max_salary,omitempty"`
	CategoriesCount  int        `json:"categories_count"`
	JobsWithMetadata int        `json:"jobs_with_metadata"`
	AvgTagsPerJob    *float64   `json:"avg_tags_per_job,omitempty"`
	LatestJob        *time.Time `json:"latest_job,omitempty"`
	FirstJob         *time.Time `json:"first_job,omitempty"`
	HoursActive      *float64   `json:"hours_active,omitempty"`
}

// GeographyData shows geographic distribution of jobs
type GeographyData struct {
	Country string         `json:"country"`
	Sources map[string]int `json:"sources"`
}

// AnalyticsResponse combines all analytics data
type AnalyticsResponse struct {
	Summary   AnalyticsSummary  `json:"summary"`
	Sources   []SourceAnalytics `json:"sources"`
	Geography []GeographyData   `json:"geography"`
}
