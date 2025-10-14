package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"openjobs/pkg/models"

	"github.com/lib/pq"
)

// JobStore handles job data operations
type JobStore struct {
	db *sql.DB
}

// NewJobStore creates a new job store
func NewJobStore(db *sql.DB) *JobStore {
	return &JobStore{db: db}
}

// CreateJob inserts a new job into the database
func (js *JobStore) CreateJob(job *models.JobPost) error {
	query := `
		INSERT INTO job_posts (id, title, company, description, location, salary,
			employment_type, experience_level, posted_date, expires_date,
			requirements, benefits, fields)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	fieldsJSON, err := json.Marshal(job.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}

	_, err = js.db.Exec(query, job.ID, job.Title, job.Company, job.Description,
		job.Location, job.Salary, job.EmploymentType, job.ExperienceLevel,
		job.PostedDate, job.ExpiresDate, pq.Array(job.Requirements),
		pq.Array(job.Benefits), fieldsJSON)

	return err
}

// GetJob retrieves a job by ID
func (js *JobStore) GetJob(id string) (*models.JobPost, error) {
	query := `
		SELECT id, title, company, description, location, salary,
			employment_type, experience_level, posted_date, expires_date,
			requirements, benefits, fields
		FROM job_posts WHERE id = $1`

	var job models.JobPost
	var fieldsJSON []byte

	err := js.db.QueryRow(query, id).Scan(
		&job.ID, &job.Title, &job.Company, &job.Description,
		&job.Location, &job.Salary, &job.EmploymentType, &job.ExperienceLevel,
		&job.PostedDate, &job.ExpiresDate, pq.Array(&job.Requirements),
		pq.Array(&job.Benefits), &fieldsJSON)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fieldsJSON, &job.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
	}

	return &job, nil
}

// GetAllJobs retrieves all jobs with optional filtering
func (js *JobStore) GetAllJobs(limit, offset int) ([]*models.JobPost, error) {
	query := `
		SELECT id, title, company, description, location, salary,
			employment_type, experience_level, posted_date, expires_date,
			requirements, benefits, fields
		FROM job_posts
		ORDER BY posted_date DESC
		LIMIT $1 OFFSET $2`

	rows, err := js.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.JobPost
	for rows.Next() {
		var job models.JobPost
		var fieldsJSON []byte

		err := rows.Scan(
			&job.ID, &job.Title, &job.Company, &job.Description,
			&job.Location, &job.Salary, &job.EmploymentType, &job.ExperienceLevel,
			&job.PostedDate, &job.ExpiresDate, pq.Array(&job.Requirements),
			pq.Array(&job.Benefits), &fieldsJSON)

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(fieldsJSON, &job.Fields)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// UpdateJob updates an existing job
func (js *JobStore) UpdateJob(job *models.JobPost) error {
	query := `
		UPDATE job_posts
		SET title = $2, company = $3, description = $4, location = $5,
			salary = $6, employment_type = $7, experience_level = $8,
			expires_date = $9, requirements = $10, benefits = $11, fields = $12
		WHERE id = $1`

	fieldsJSON, err := json.Marshal(job.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}

	_, err = js.db.Exec(query, job.ID, job.Title, job.Company, job.Description,
		job.Location, job.Salary, job.EmploymentType, job.ExperienceLevel,
		job.ExpiresDate, pq.Array(job.Requirements), pq.Array(job.Benefits), fieldsJSON)

	return err
}

// DeleteJob removes a job from the database
func (js *JobStore) DeleteJob(id string) error {
	query := `DELETE FROM job_posts WHERE id = $1`
	_, err := js.db.Exec(query, id)
	return err
}