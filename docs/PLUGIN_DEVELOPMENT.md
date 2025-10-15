# OpenJobs Plugin Architecture & Development Guide

This document explains the **properly decoupled plugin architecture** of OpenJobs and provides guidelines for creating new plugins that can integrate seamlessly without modifying core code.

## Plugin Architecture Overview

OpenJobs implements a **clean plugin architecture** using Go interfaces and a registry pattern. This ensures:

- **🔌 Zero Coupling**: Core system doesn't know about specific plugins
- **📦 Easy Extension**: Add new connectors without touching core code
- **🏗️ Clean Interfaces**: Standard contracts for all plugins
- **🔄 Dynamic Loading**: Plugins register themselves at startup

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Core System   │    │  PluginRegistry │    │    Connectors   │
│                 │◄──►│                 │◄──►│                 │
│  • API Server   │    │ • Register()    │    │ • Arbetsförmed  │
│  • Scheduler    │    │ • GetConnector()│    │ • EURES/Adzuna │
│  • Storage      │    │ • ListAll()     │    │ • Custom...     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Plugin Overview

Plugins extend the platform's capabilities by connecting to external job sources. All plugins are implemented as Go modules and registered at runtime within the main application binary — no standalone binaries or external configuration files are used.

## Plugin Structure

A typical plugin consists of:

```
connectors/arbetsformedlingen/
├── connector.go         # Implements PluginConnector interface
├── README.md            # Plugin documentation
└── go.mod               # Go module dependencies
```

All plugins must be located under the `connectors/` directory and compiled into the main binary. There is no external plugin loading mechanism.

## Plugin Types

### 1. Data Source Plugins
Connect to external job platforms like Arbetsförmedlingen or EURES via their public APIs.

## Development Process

### Step 1: Implement the PluginConnector Interface
Create a new Go file in the `connectors/` directory that implements the `PluginConnector` interface from `pkg/models/plugin.go`:

```go
package arbetsformedlingen

import (
	"openjobs/pkg/models"
	"openjobs/pkg/storage"
)

type ArbetsformedlingenConnector struct {
	store *storage.JobStore
}

func NewArbetsformedlingenConnector(store *storage.JobStore) *ArbetsformedlingenConnector {
	return &ArbetsformedlingenConnector{
		store: store,
	}
}

func (ac *ArbetsformedlingenConnector) GetID() string {
	return "arbetsformedlingen"
}

func (ac *ArbetsformedlingenConnector) GetName() string {
	return "Arbetsförmedlingen Connector"
}

func (ac *ArbetsformedlingenConnector) FetchJobs() ([]models.JobPost, error) {
	// Implement job fetching logic here
	// Use environment variables for API keys (e.g., os.Getenv("ADZUNA_APP_ID"))
	// Always fallback to demo data if keys are missing
}

func (ac *ArbetsformedlingenConnector) SyncJobs() error {
	// Fetch jobs and store them using ac.store.CreateJob(&job)
	// Handle deduplication by checking if job.ID already exists
}
```

### Step 2: Register the Connector in the Scheduler
Open `internal/scheduler/scheduler.go` and add your connector to the registry in the `NewScheduler()` function:

```go
func NewScheduler(store *storage.JobStore) *Scheduler {
	registry := NewPluginRegistry()
	
	// Register your connector here
	registry.Register(arbetsformedlingen.NewArbetsformedlingenConnector(store))
	registry.Register(eures.NewEURESConnector(store))
	
	return &Scheduler{
		store:      store,
		registry:   registry,
		ticker:     time.NewTicker(6 * time.Hour),
	}
}
```

### Step 3: Configure Environment Variables
Use environment variables for sensitive configuration (API keys, URLs). Add your variables to `.env.example`:

```bash
# Example for EURES connector
ADZUNA_APP_ID=your-app-id-here
ADZUNA_APP_KEY=your-app-key-here

# Example for Arbetsförmedlingen connector (if needed)
ARBETSFORMEDLINGEN_API_KEY=your-api-token-here
```

Then load them in your connector using `os.Getenv()`:

```go
appID := os.Getenv("ADZUNA_APP_ID")
if appID == "" {
    log.Println("⚠️  Adzuna credentials not configured, using demo data")
    return fetchDemoJobs()
}
```

### Step 4: Write Tests
Include unit tests for your connector logic:

```go
func TestFetchJobs(t *testing.T) {
	store := &mockJobStore{}
	connector := NewArbetsformedlingenConnector(store)
	
	jobs, err := connector.FetchJobs()
	assert.NoError(t, err)
	assert.Greater(t, len(jobs), 0)
}
```

## Integration with Job Platform

### Configuration Management
All configuration is done via environment variables. No `config.json` files are used. Use `.env.example` as a template for required variables.

### Data Transformation
Transform external job data to the `models.JobPost` format. Ensure all required fields are populated and use the `Fields` map for source-specific metadata:

```go
job := models.JobPost{
	ID:              fmt.Sprintf("af-%s", afJob.ID),
	Title:           afJob.Headline,
	Company:         afJob.Employer.Name,
	Description:     extractDescription(afJob),
	Location:        formatLocation(afJob),
	Salary:          afJob.SalaryDescription,
	EmploymentType:  mapEmploymentType(afJob.EmploymentType.ConceptLabel),
	ExperienceLevel: mapExperienceLevel(afJob.ExperienceRequired),
	PostedDate:      parseAFDate(afJob.PublicationDate),
	ExpiresDate:     parseAFDate(afJob.LastApplicationDate),
	Fields: map[string]interface{}{
		"source":       "arbetsformedlingen",
		"source_url":   extractURL(afJob),
		"original_id":  afJob.ID,
		"country":      afJob.WorkplaceAddress.Country,
		"region":       afJob.WorkplaceAddress.Region,
		"municipality": afJob.WorkplaceAddress.Municipality,
		"connector":    "arbetsformedlingen",
		"fetched_at":   time.Now(),
	},
}
```

## Testing Your Plugin

### Unit Tests
Include tests for `FetchJobs()` and `SyncJobs()`:

```go
func TestSyncJobs(t *testing.T) {
	store := &mockJobStore{}
	connector := NewArbetsformedlingenConnector(store)
	
	err := connector.SyncJobs()
	assert.NoError(t, err)
	assert.Equal(t, 1, store.CreatedCount) // Verify job was stored
}
```

## Deployment

### Build and Run
Plugins are compiled into the main binary. Build the entire application:

```bash
cd cmd/openjobs
go build -o openjobs main.go
./openjobs
```

### Production Deployment
1. Build the application as a Docker container
2. Push to container registry
3. Set environment variables in deployment (e.g., Kubernetes secrets or Docker env)
4. No plugin registration step is needed — connectors are compiled in

## Best Practices

1. **Error Handling**: Always implement robust error handling for HTTP requests and JSON parsing
2. **Rate Limiting**: Respect API rate limits of external services (add delays if needed)
3. **Logging**: Use `fmt.Println()` or `log.Printf()` for debugging during sync
4. **Security**: Protect sensitive credentials with environment variables — never hardcode
5. **Validation**: Validate all incoming data before storing in database
6. **Documentation**: Document your connector’s API endpoints and required env vars in README.md
7. **Testing**: Include unit tests for core logic — especially data transformation
8. **Fallbacks**: Always provide demo data fallbacks when API keys are missing

## Common Issues

### Authentication Problems
Ensure API tokens and credentials are properly configured in `.env` and loaded via `os.Getenv()`.

### Data Mapping Issues
Verify that field mappings between external sources and `models.JobPost` are correct. Use the `Fields` map for source-specific metadata.

### Performance Bottlenecks
Implement efficient data fetching and processing to handle large job datasets. Use pagination and batch processing if needed.

## Support Resources

- GitHub Repository: https://github.com/yourusername/job-platform
- Documentation: https://github.com/yourusername/job-platform/wiki
- Community Forum: https://github.com/yourusername/job-platform/discussions

## Plugin Types

### 1. Data Source Plugins
Connect to external job platforms like Arbetsförmedlingen or EURES via their public APIs.

## Development Process

### Step 1: Implement the PluginConnector Interface
Create a new Go file in the `connectors/` directory that implements the `PluginConnector` interface from `pkg/models/plugin.go`:

```go
package arbetsformedlingen

import (
	"openjobs/pkg/models"
	"openjobs/pkg/storage"
)

type ArbetsformedlingenConnector struct {
	store *storage.JobStore
}

func NewArbetsformedlingenConnector(store *storage.JobStore) *ArbetsformedlingenConnector {
	return &ArbetsformedlingenConnector{
		store: store,
	}
}

func (ac *ArbetsformedlingenConnector) GetID() string {
	return "arbetsformedlingen"
}

func (ac *ArbetsformedlingenConnector) GetName() string {
	return "Arbetsförmedlingen Connector"
}

func (ac *ArbetsformedlingenConnector) FetchJobs() ([]models.JobPost, error) {
	// Implement job fetching logic here
	// Use environment variables for API keys (e.g., os.Getenv("ADZUNA_APP_ID"))
	// Always fallback to demo data if keys are missing
}

func (ac *ArbetsformedlingenConnector) SyncJobs() error {
	// Fetch jobs and store them using ac.store.CreateJob(&job)
	// Handle deduplication by checking if job.ID already exists
}
```

### Step 2: Register the Connector in the Scheduler
Open `internal/scheduler/scheduler.go` and add your connector to the registry in the `NewScheduler()` function:

```go
func NewScheduler(store *storage.JobStore) *Scheduler {
	registry := NewPluginRegistry()
	
	// Register your connector here
	registry.Register(arbetsformedlingen.NewArbetsformedlingenConnector(store))
	registry.Register(eures.NewEURESConnector(store))
	
	return &Scheduler{
		store:      store,
		registry:   registry,
		ticker:     time.NewTicker(6 * time.Hour),
	}
}
```

### Step 3: Configure Environment Variables
Use environment variables for sensitive configuration (API keys, URLs). Add your variables to `.env.example`:

```bash
# Example for EURES connector
ADZUNA_APP_ID=your-app-id-here
ADZUNA_APP_KEY=your-app-key-here

# Example for Arbetsförmedlingen connector (if needed)
ARBETSFORMEDLINGEN_API_KEY=your-api-token-here
```

Then load them in your connector using `os.Getenv()`:

```go
appID := os.Getenv("ADZUNA_APP_ID")
if appID == "" {
    log.Println("⚠️  Adzuna credentials not configured, using demo data")
    return fetchDemoJobs()
}
```

### Step 4: Write Tests
Include unit tests for your connector logic:

```go
func TestFetchJobs(t *testing.T) {
	store := &mockJobStore{}
	connector := NewArbetsformedlingenConnector(store)
	
	jobs, err := connector.FetchJobs()
	assert.NoError(t, err)
	assert.Greater(t, len(jobs), 0)
}
```

## Integration with Job Platform

### Configuration Management
All configuration is done via environment variables. No `config.json` files are used. Use `.env.example` as a template for required variables.

### Data Transformation
Transform external job data to the `models.JobPost` format. Ensure all required fields are populated and use the `Fields` map for source-specific metadata:

```go
job := models.JobPost{
	ID:              fmt.Sprintf("af-%s", afJob.ID),
	Title:           afJob.Headline,
	Company:         afJob.Employer.Name,
	Description:     extractDescription(afJob),
	Location:        formatLocation(afJob),
	Salary:          afJob.SalaryDescription,
	EmploymentType:  mapEmploymentType(afJob.EmploymentType.ConceptLabel),
	ExperienceLevel: mapExperienceLevel(afJob.ExperienceRequired),
	PostedDate:      parseAFDate(afJob.PublicationDate),
	ExpiresDate:     parseAFDate(afJob.LastApplicationDate),
	Fields: map[string]interface{}{
		"source":       "arbetsformedlingen",
		"source_url":   extractURL(afJob),
		"original_id":  afJob.ID,
		"country":      afJob.WorkplaceAddress.Country,
		"region":       afJob.WorkplaceAddress.Region,
		"municipality": afJob.WorkplaceAddress.Municipality,
		"connector":    "arbetsformedlingen",
		"fetched_at":   time.Now(),
	},
}
```

## Configuration Management

### Plugin Settings
Plugins should support configurable settings:
```json
{
  "sync_frequency": "hourly",
  "max_jobs_per_sync": 100,
  "retry_attempts": 3,
  "log_level": "info"
}
```

### Environment Variables
```bash
export PLUGIN_NAME="Arbetsförmedlingen Connector"
export PLUGIN_TOKEN="your-secret-token"
export SYNC_INTERVAL="3600"
```

## Testing Your Plugin

### Unit Tests
```go
func TestTransformJob(t *testing.T) {
    plugin := &Plugin{}
    
    sourceJob := map[string]interface{}{
        "title": "Software Engineer",
        "company": "Tech Corp",
        "description": "Looking for experienced engineers",
        "location": "Stockholm",
        "salary": "SEK 40,000-60,000",
        "employment_type": "Full-time",
        "experience_level": "Mid-level",
    }
    
    transformed := plugin.TransformJob(sourceJob)
    
    assert.Equal(t, "Software Engineer", transformed.Title)
    assert.Equal(t, "Tech Corp", transformed.Company)
    // ... additional assertions
}
```

## Deployment

### Build and Run
Plugins are compiled into the main binary. Build the entire application:

```bash
cd cmd/openjobs
go build -o openjobs main.go
./openjobs
```

### Production Deployment
1. Build the application as a Docker container
2. Push to container registry
3. Set environment variables in deployment (e.g., Kubernetes secrets or Docker env)
4. No plugin registration step is needed — connectors are compiled in

## Best Practices

1. **Error Handling**: Always implement robust error handling
2. **Rate Limiting**: Respect API rate limits of external services
3. **Logging**: Include comprehensive logging for debugging
4. **Security**: Protect sensitive credentials with environment variables
5. **Validation**: Validate all incoming data before processing
6. **Documentation**: Provide clear documentation for users
7. **Testing**: Include unit and integration tests

## Common Issues

### Authentication Problems
Ensure API tokens and credentials are properly configured in environment variables.

### Data Mapping Issues
Verify that field mappings between external sources and platform format are correct.

### Performance Bottlenecks
Implement efficient data fetching and processing to handle large job datasets.

## Support Resources

- GitHub Repository: https://github.com/yourusername/job-platform
- Documentation: https://github.com/yourusername/job-platform/wiki
- Community Forum: https://github.com/yourusername/job-platform/discussions
