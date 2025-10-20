package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPPluginConnector implements PluginConnector interface via HTTP calls
type HTTPPluginConnector struct {
	pluginID   string
	pluginName string
	baseURL    string
	httpClient *http.Client
}

// HTTPPluginResponse represents the response from plugin HTTP endpoints
type HTTPPluginResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// NewHTTPPluginConnector creates a new HTTP-based plugin connector
func NewHTTPPluginConnector(id, name, url string) *HTTPPluginConnector {
	return &HTTPPluginConnector{
		pluginID:   id,
		pluginName: name,
		baseURL:    strings.TrimSuffix(url, "/"),
		httpClient: &http.Client{Timeout: 6 * time.Minute}, // 6 minutes for Chrome scraping
	}
}

// GetID returns the plugin ID
func (h *HTTPPluginConnector) GetID() string {
	return h.pluginID
}

// GetName returns the plugin name
func (h *HTTPPluginConnector) GetName() string {
	return h.pluginName
}

// FetchJobs fetches jobs via HTTP from the plugin service
func (h *HTTPPluginConnector) FetchJobs() ([]JobPost, error) {
	url := h.baseURL + "/jobs"

	resp, err := h.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs from plugin %s: %w", h.pluginName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("plugin %s returned status %d: %s", h.pluginName, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from plugin %s: %w", h.pluginName, err)
	}

	var response HTTPPluginResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response from plugin %s: %w", h.pluginName, err)
	}

	if !response.Success {
		return nil, fmt.Errorf("plugin %s error: %s", h.pluginName, response.Error)
	}

	// Convert interface{} data to []JobPost
	jobsData, ok := response.Data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("plugin %s returned invalid data format", h.pluginName)
	}

	var jobs []JobPost
	for _, jobData := range jobsData {
		jobMap, ok := jobData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("plugin %s returned invalid job data format", h.pluginName)
		}

		job, err := h.mapToJobPost(jobMap)
		if err != nil {
			return nil, fmt.Errorf("failed to map job data from plugin %s: %w", h.pluginName, err)
		}

		jobs = append(jobs, *job)
	}

	return jobs, nil
}

// SyncJobs triggers job synchronization via HTTP
func (h *HTTPPluginConnector) SyncJobs() error {
	url := h.baseURL + "/sync"

	resp, err := h.httpClient.Post(url, "application/json", strings.NewReader("{}"))
	if err != nil {
		return fmt.Errorf("failed to sync jobs with plugin %s: %w", h.pluginName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("plugin %s sync returned status %d: %s", h.pluginName, resp.StatusCode, string(body))
	}

	return nil
}

// mapToJobPost converts map[string]interface{} to JobPost struct
func (h *HTTPPluginConnector) mapToJobPost(data map[string]interface{}) (*JobPost, error) {
	job := &JobPost{}

	// Basic field mapping
	if id, ok := data["id"].(string); ok {
		job.ID = id
	}
	if title, ok := data["title"].(string); ok {
		job.Title = title
	}
	if company, ok := data["company"].(string); ok {
		job.Company = company
	}
	if location, ok := data["location"].(string); ok {
		job.Location = location
	}
	if salary, ok := data["salary"].(string); ok {
		job.Salary = salary
	}
	if employmentType, ok := data["employment_type"].(string); ok {
		job.EmploymentType = employmentType
	}
	if experienceLevel, ok := data["experience_level"].(string); ok {
		job.ExperienceLevel = experienceLevel
	}
	if description, ok := data["description"].(string); ok {
		job.Description = description
	}

	// Handle timestamp fields
	if postedDate, exists := data["posted_date"]; exists {
		if postedDateStr, ok := postedDate.(string); ok {
			if t, err := time.Parse(time.RFC3339, postedDateStr); err == nil {
				job.PostedDate = t
			}
		}
	}

	if expiresDate, exists := data["expires_date"]; exists {
		if expiresDateStr, ok := expiresDate.(string); ok {
			if t, err := time.Parse(time.RFC3339, expiresDateStr); err == nil {
				job.ExpiresDate = t
			}
		}
	}

	// Handle requirements (string array)
	if reqs, exists := data["requirements"]; exists {
		if reqsSlice, ok := reqs.([]interface{}); ok {
			for _, req := range reqsSlice {
				if reqStr, ok := req.(string); ok {
					job.Requirements = append(job.Requirements, reqStr)
				}
			}
		}
	}

	// Handle benefits (string array)
	if benefits, exists := data["benefits"]; exists {
		if benefitsSlice, ok := benefits.([]interface{}); ok {
			for _, benefit := range benefitsSlice {
				if benefitStr, ok := benefit.(string); ok {
					job.Benefits = append(job.Benefits, benefitStr)
				}
			}
		}
	}

	// Handle fields (map)
	if fields, exists := data["fields"]; exists {
		if fieldsMap, ok := fields.(map[string]interface{}); ok {
			job.Fields = fieldsMap
		}
	}

	return job, nil
}
