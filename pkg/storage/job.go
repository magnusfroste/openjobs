package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"openjobs/pkg/models"
)

// JobStore handles job data operations
type JobStore struct {
	supabaseURL    string
	supabaseKey    string
	serviceRoleKey string
	httpClient     *http.Client
}

// NewJobStore creates a new job store
func NewJobStore() *JobStore {
	return &JobStore{
		supabaseURL:    os.Getenv("SUPABASE_URL"),
		supabaseKey:    os.Getenv("SUPABASE_ANON_KEY"),
		serviceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		httpClient:     &http.Client{},
	}
}

// CreateJob inserts a new job into Supabase
func (js *JobStore) CreateJob(job *models.JobPost) error {
	fmt.Printf("📝 Attempting to create job: %s (ID: %s)\n", job.Title, job.ID)

	jobJSON, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/job_posts", js.supabaseURL)
	fmt.Printf("   POST to: %s\n", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jobJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use service_role key for write operations (bypasses RLS after API key validation)
	key := js.serviceRoleKey
	if key == "" {
		key = js.supabaseKey // Fallback to anon key if service_role not set
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	req.Header.Set("apikey", key)
	req.Header.Set("Prefer", "return=representation")

	resp, err := js.httpClient.Do(req)
	if err != nil {
		fmt.Printf("   ❌ HTTP request failed: %v\n", err)
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		fmt.Printf("   ❌ Supabase error %d: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("supabase error %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("   ✅ Job created successfully (status: %d)\n", resp.StatusCode)
	return nil
}

// GetJob retrieves a job by ID from Supabase
func (js *JobStore) GetJob(id string) (*models.JobPost, error) {
	url := fmt.Sprintf("%s/rest/v1/job_posts?id=eq.%s", js.supabaseURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("sql: no rows in result set")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var jobs []models.JobPost
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(jobs) == 0 {
		return nil, fmt.Errorf("sql: no rows in result set")
	}

	return &jobs[0], nil
}

// GetAllJobs retrieves all jobs with optional filtering from Supabase
func (js *JobStore) GetAllJobs(limit, offset int) ([]*models.JobPost, error) {
	url := fmt.Sprintf("%s/rest/v1/job_posts?select=*&is_active=eq.true&order=posted_date.desc&limit=%d&offset=%d", js.supabaseURL, limit, offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("supabase error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var jobs []*models.JobPost
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return jobs, nil
}

// GetJobsAfter retrieves jobs created after a specific timestamp (for incremental sync)
func (js *JobStore) GetJobsAfter(timestamp string, limit, offset int) ([]*models.JobPost, error) {
	// Use created_at for incremental sync - this is when OpenJobs added the job to database
	// NOT posted_date which is when the employer originally posted it (can be old)
	url := fmt.Sprintf("%s/rest/v1/job_posts?select=*&is_active=eq.true&created_at=gte.%s&order=created_at.desc&limit=%d&offset=%d",
		js.supabaseURL, timestamp, limit, offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("supabase error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var jobs []*models.JobPost
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return jobs, nil
}

// UpdateJob updates an existing job in Supabase
func (js *JobStore) UpdateJob(job *models.JobPost) error {
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/job_posts?id=eq.%s", js.supabaseURL, job.ID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jobJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use service_role key for write operations (bypasses RLS)
	key := js.serviceRoleKey
	if key == "" {
		key = js.supabaseKey
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	req.Header.Set("apikey", key)

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteJob removes a job from Supabase
func (js *JobStore) DeleteJob(id string) error {
	url := fmt.Sprintf("%s/rest/v1/job_posts?id=eq.%s", js.supabaseURL, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use service_role key for write operations (bypasses RLS)
	key := js.serviceRoleKey
	if key == "" {
		key = js.supabaseKey
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	req.Header.Set("apikey", key)

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// LogSync creates a sync log entry
func (js *JobStore) LogSync(log *models.SyncLog) error {
	logJSON, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal sync log: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/sync_logs", js.supabaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(logJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)
	req.Header.Set("Prefer", "return=representation")

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase error %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("📊 Sync log created: %s - Fetched: %d, Inserted: %d, Duplicates: %d\n",
		log.ConnectorName, log.JobsFetched, log.JobsInserted, log.JobsDuplicates)
	return nil
}

// GetRecentSyncLogs retrieves recent sync logs
func (js *JobStore) GetRecentSyncLogs(limit int) ([]models.SyncLog, error) {
	url := fmt.Sprintf("%s/rest/v1/sync_logs?order=started_at.desc&limit=%d", js.supabaseURL, limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("supabase error %d: %s", resp.StatusCode, string(body))
	}

	var logs []models.SyncLog
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return logs, nil
}

// GetTotalJobCount returns the total number of active jobs in the database
func (js *JobStore) GetTotalJobCount() (int, error) {
	url := fmt.Sprintf("%s/rest/v1/job_posts?select=count&is_active=eq.true", js.supabaseURL)
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)
	req.Header.Set("Prefer", "count=exact")

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("supabase error %d", resp.StatusCode)
	}

	// Parse Content-Range header: "0-24/145" -> 145
	contentRange := resp.Header.Get("Content-Range")
	if contentRange == "" {
		return 0, nil
	}

	// Extract total from "0-24/145"
	parts := strings.Split(contentRange, "/")
	if len(parts) != 2 {
		return 0, nil
	}

	total, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse total count: %w", err)
	}

	return total, nil
}

// GetRemoteJobCount returns the number of active remote jobs in the database
func (js *JobStore) GetRemoteJobCount() (int, error) {
	url := fmt.Sprintf("%s/rest/v1/job_posts?select=count&is_active=eq.true&is_remote=eq.true", js.supabaseURL)
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)
	req.Header.Set("Prefer", "count=exact")

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("supabase error %d", resp.StatusCode)
	}

	// Parse Content-Range header
	contentRange := resp.Header.Get("Content-Range")
	if contentRange == "" {
		return 0, nil
	}

	parts := strings.Split(contentRange, "/")
	if len(parts) != 2 {
		return 0, nil
	}

	total, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse total count: %w", err)
	}

	return total, nil
}

// GetMostRecentJob retrieves the most recent job for a given connector (by ID prefix)
// Used for incremental sync - finds the last job posted to determine where to continue
func (js *JobStore) GetMostRecentJob(idPrefix string) (*models.JobPost, error) {
	// Query for most recent job with matching ID prefix, ordered by created_at
	url := fmt.Sprintf("%s/rest/v1/job_posts?select=*&id=like.%s*&order=created_at.desc&limit=1",
		js.supabaseURL, idPrefix)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)

	resp, err := js.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("supabase error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var jobs []models.JobPost
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(jobs) == 0 {
		return nil, fmt.Errorf("no jobs found with prefix %s", idPrefix)
	}

	return &jobs[0], nil
}
