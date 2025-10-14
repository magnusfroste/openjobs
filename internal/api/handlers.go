package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"jobplatform/pkg/models"
	"jobplatform/pkg/storage"

	"github.com/google/uuid"
)

// Server holds the HTTP server dependencies
type Server struct {
	jobStore *storage.JobStore
}

// NewServer creates a new server instance
func NewServer(jobStore *storage.JobStore) *Server {
	return &Server{jobStore: jobStore}
}

// CreateJob handles POST /jobs
func (s *Server) CreateJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var job models.JobPost
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	// Set posted date if not provided
	if job.PostedDate.IsZero() {
		job.PostedDate = time.Now()
	}

	if err := s.jobStore.CreateJob(&job); err != nil {
		http.Error(w, `{"success": false, "message": "Failed to create job"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    job,
		Message: "Job created successfully",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetAllJobs handles GET /jobs
func (s *Server) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	limit := 20 // default
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	jobs, err := s.jobStore.GetAllJobs(limit, offset)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Failed to retrieve jobs"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    jobs,
	}

	json.NewEncoder(w).Encode(response)
}

// GetJobByID handles GET /jobs/{id}
func (s *Server) GetJobByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/jobs/"):] // Extract ID from path
	if id == "" {
		http.Error(w, `{"success": false, "message": "Job ID required"}`, http.StatusBadRequest)
		return
	}

	job, err := s.jobStore.GetJob(id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, `{"success": false, "message": "Job not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"success": false, "message": "Failed to retrieve job"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    job,
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateJob handles PUT /jobs/{id}
func (s *Server) UpdateJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/jobs/"):] // Extract ID from path
	if id == "" {
		http.Error(w, `{"success": false, "message": "Job ID required"}`, http.StatusBadRequest)
		return
	}

	var job models.JobPost
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	job.ID = id // Ensure ID matches path

	if err := s.jobStore.UpdateJob(&job); err != nil {
		http.Error(w, `{"success": false, "message": "Failed to update job"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    job,
		Message: "Job updated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// DeleteJob handles DELETE /jobs/{id}
func (s *Server) DeleteJob(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/jobs/"):] // Extract ID from path
	if id == "" {
		http.Error(w, `{"success": false, "message": "Job ID required"}`, http.StatusBadRequest)
		return
	}

	if err := s.jobStore.DeleteJob(id); err != nil {
		http.Error(w, `{"success": false, "message": "Failed to delete job"}`, http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Job deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
}