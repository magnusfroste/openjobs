# Job Platform Plugin Development Guide

This document provides guidelines for creating plugins that can integrate with the Job Platform.

## Plugin Overview

Plugins extend the platform's capabilities by connecting to external job sources, transforming data formats, or adding new functionality.

## Plugin Structure

A typical plugin consists of:

```
my-job-plugin/
├── main.go              # Plugin entry point
├── config.json          # Plugin configuration
├── handlers/            # Plugin-specific handlers
│   ├── api.go           # API integration handlers
│   └── data.go          # Data transformation logic
├── README.md            # Plugin documentation
└── go.mod               # Go module dependencies
```

## Plugin Types

### 1. Data Source Plugins
Connect to external job platforms like Arbetsförmedlingen or LinkedIn.

### 2. Transformation Plugins
Convert data between different formats (XML, JSON, CSV).

### 3. Synchronization Plugins
Handle periodic data updates and real-time feeds.

## Development Process

### Step 1: Create Plugin Module
```bash
mkdir my-job-plugin
cd my-job-plugin
go mod init my-job-plugin
```

### Step 2: Define Configuration
Create `config.json`:
```json
{
  "name": "Arbetsförmedlingen Connector",
  "version": "1.0.0",
  "description": "Connects to Arbetsförmedlingen job data",
  "source_url": "https://api.arbetsformedlingen.se",
  "authentication": {
    "type": "bearer_token",
    "token": "your-api-token"
  }
}
```

### Step 3: Implement Plugin Logic
In `main.go`:
```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "your-platform/pkg/models"
)

type Plugin struct {
    Name        string
    Version     string
    Config      PluginConfig
}

type PluginConfig struct {
    Name        string `json:"name"`
    SourceURL   string `json:"source_url"`
    Auth        Auth   `json:"authentication"`
}

type Auth struct {
    Type  string `json:"type"`
    Token string `json:"token"`
}

func (p *Plugin) FetchJobs() ([]models.JobPost, error) {
    // Implementation to fetch jobs from external source
    client := &http.Client{Timeout: 30 * time.Second}
    req, err := http.NewRequest("GET", p.Config.SourceURL, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.Config.Auth.Token))
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Process response and convert to JobPost format
    var jobs []models.JobPost
    // ... conversion logic
    
    return jobs, nil
}

func main() {
    // Plugin initialization and registration
    plugin := &Plugin{
        Name:    "Arbetsförmedlingen Connector",
        Version: "1.0.0",
    }
    
    // Register with platform
    fmt.Println("Plugin registered successfully")
}
```

## Integration with Job Platform

### API Integration
Plugins can make HTTP requests to the platform's API:
```go
func (p *Plugin) SubmitJobs(jobs []models.JobPost) error {
    url := "http://localhost:8080/jobs/create"
    
    for _, job := range jobs {
        jsonData, err := json.Marshal(job)
        if err != nil {
            return err
        }
        
        resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
            return err
        }
        defer resp.Body.Close()
    }
    
    return nil
}
```

### Data Transformation
Convert external job data to platform format:
```go
func (p *Plugin) TransformJob(sourceJob map[string]interface{}) models.JobPost {
    job := models.JobPost{
        Title:           sourceJob["title"].(string),
        Company:         sourceJob["company"].(string),
        Description:     sourceJob["description"].(string),
        Location:        sourceJob["location"].(string),
        Salary:          sourceJob["salary"].(string),
        EmploymentType:  sourceJob["employment_type"].(string),
        ExperienceLevel: sourceJob["experience_level"].(string),
        PostedDate:      time.Now(),
        ExpiresDate:     time.Now().AddDate(0, 6, 0),
    }
    
    // Handle additional fields
    if fields, ok := sourceJob["fields"].(map[string]interface{}); ok {
        job.Fields = fields
    }
    
    return job
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

### Local Testing
```bash
# Build plugin
go build -o my-plugin main.go

# Run plugin
./my-plugin
```

### Production Deployment
1. Package plugin as Docker container
2. Push to container registry
3. Configure in platform's plugin manager
4. Set environment variables
5. Schedule periodic execution if needed

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