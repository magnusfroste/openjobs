package models

import (
	"time"

	"github.com/lib/pq"
)

// JobPost represents a job posting with flexible attributes
type JobPost struct {
	ID              string                 `json:"id" db:"id"`
	Title           string                 `json:"title" db:"title"`
	Company         string                 `json:"company" db:"company"`
	Description     string                 `json:"description" db:"description"`
	Location        string                 `json:"location" db:"location"`
	Salary          string                 `json:"salary" db:"salary"`
	SalaryMin       *int                   `json:"salary_min,omitempty" db:"salary_min"`
	SalaryMax       *int                   `json:"salary_max,omitempty" db:"salary_max"`
	SalaryCurrency  string                 `json:"salary_currency,omitempty" db:"salary_currency"`
	IsRemote        bool                   `json:"is_remote" db:"is_remote"`
	URL             string                 `json:"url,omitempty" db:"url"`
	EmploymentType  string                 `json:"employment_type" db:"employment_type"`
	ExperienceLevel string                 `json:"experience_level" db:"experience_level"`
	PostedDate      time.Time              `json:"posted_date" db:"posted_date"`
	ExpiresDate     time.Time              `json:"expires_date" db:"expires_date"`
	Requirements    []string               `json:"requirements" db:"requirements"`
	Benefits        []string               `json:"benefits" db:"benefits"`
	Fields          map[string]interface{} `json:"fields" db:"fields"`
}

// JobPostTraditional represents a job posting with fixed schema
type JobPostTraditional struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Company         string         `json:"company"`
	Description     string         `json:"description"`
	Location        string         `json:"location"`
	Salary          string         `json:"salary"`
	EmploymentType  string         `json:"employment_type"`
	ExperienceLevel string         `json:"experience_level"`
	PostedDate      time.Time      `json:"posted_date"`
	ExpiresDate     time.Time      `json:"expires_date"`
	Requirements    pq.StringArray `json:"requirements"`
	Benefits        pq.StringArray `json:"benefits"`
}

// PluginInfo represents plugin metadata
type PluginInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Source      string    `json:"source"`
	Status      string    `json:"status"`
	LastRun     time.Time `json:"last_run"`
	NextRun     time.Time `json:"next_run"`
	Description string    `json:"description"`
}

// APIResponse represents a job listing response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// SyncLog represents a connector sync operation log
type SyncLog struct {
	ID             string    `json:"id,omitempty" db:"id"`
	ConnectorName  string    `json:"connector_name" db:"connector_name"`
	StartedAt      time.Time `json:"started_at" db:"started_at"`
	CompletedAt    time.Time `json:"completed_at" db:"completed_at"`
	JobsFetched    int       `json:"jobs_fetched" db:"jobs_fetched"`
	JobsInserted   int       `json:"jobs_inserted" db:"jobs_inserted"`
	JobsDuplicates int       `json:"jobs_duplicates" db:"jobs_duplicates"`
	Status         string    `json:"status" db:"status"` // success, error, partial
	ErrorMessage   string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt      time.Time `json:"created_at,omitempty" db:"created_at"`
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Fetched    int
	Inserted   int
	Duplicates int
	Errors     []string
}
