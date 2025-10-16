package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"openjobs/pkg/models"
)

// JobStore handles job data operations
type JobStore struct {
	supabaseURL string
	supabaseKey string
	httpClient  *http.Client
}

// NewJobStore creates a new job store
func NewJobStore() *JobStore {
	return &JobStore{
		supabaseURL: os.Getenv("SUPABASE_URL"),
		supabaseKey: os.Getenv("SUPABASE_ANON_KEY"),
		httpClient:  &http.Client{},
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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)
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
	url := fmt.Sprintf("%s/rest/v1/job_posts?select=*&order=posted_date.desc&limit=%d&offset=%d", js.supabaseURL, limit, offset)
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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)

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

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", js.supabaseKey))
	req.Header.Set("apikey", js.supabaseKey)

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
