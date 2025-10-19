# Jooble Job Aggregator Connector

Job aggregator connector for OpenJobs that fetches jobs from Jooble API.

## Overview

Jooble is a job search engine that aggregates job listings from multiple sources including:
- Company career pages
- Job boards (Indeed, Monster, etc.)
- LinkedIn
- Local job sites
- Recruitment agencies

This connector provides access to Swedish job listings through Jooble's free API.

## Features

- ✅ **Free API** - No cost, just requires API key
- ✅ **Multi-source aggregation** - Jobs from multiple job boards
- ✅ **Sweden coverage** - Good Swedish job market coverage
- ✅ **Multiple search queries** - developer, engineer, designer, manager, sales, marketing
- ✅ **Smart deduplication** - Removes duplicate job listings
- ✅ **Remote job detection** - Identifies remote work opportunities
- ✅ **Skills extraction** - Extracts 40+ tech skills from descriptions
- ✅ **Demo mode** - Works without API key for testing

## Configuration

### Environment Variables

```bash
# Required for production
JOOBLE_API_KEY=your_api_key_here

# Database (required)
SUPABASE_URL=your_supabase_url
SUPABASE_ANON_KEY=your_supabase_key

# Optional
PORT=8088  # Default: 8088
```

### Getting API Key

1. Visit https://jooble.org/api/about
2. Register for a free API key
3. Add to your `.env` file

## API Endpoints

### Health Check
```bash
GET /health
```

Response:
```json
{
  "status": "healthy",
  "service": "jooble-plugin",
  "version": "1.0.0"
}
```

### Trigger Sync
```bash
POST /sync
```

Response:
```json
{
  "status": "success",
  "message": "Jooble sync completed successfully"
}
```

### List Jobs
```bash
GET /jobs
```

Response:
```json
{
  "status": "success",
  "count": 250,
  "jobs": [...]
}
```

## Running Locally

### With API Key
```bash
export JOOBLE_API_KEY=your_key
export SUPABASE_URL=your_url
export SUPABASE_ANON_KEY=your_key
PORT=8088 go run cmd/plugin-jooble/main.go
```

### Demo Mode (No API Key)
```bash
export SUPABASE_URL=your_url
export SUPABASE_ANON_KEY=your_key
PORT=8088 go run cmd/plugin-jooble/main.go
```

## Docker

### Build
```bash
docker build -t openjobs-jooble -f connectors/jooble/Dockerfile .
```

### Run
```bash
docker run -p 8088:8088 \
  -e JOOBLE_API_KEY=your_key \
  -e SUPABASE_URL=your_url \
  -e SUPABASE_ANON_KEY=your_key \
  openjobs-jooble
```

## Expected Results

### Per Sync
- **Queries**: 6 (developer, engineer, designer, manager, sales, marketing)
- **Jobs per query**: ~50-100
- **Total jobs**: ~200-400
- **Unique jobs**: ~150-300 (after deduplication)
- **Sync time**: ~15-30 seconds

### Data Quality
- ✅ Job titles
- ✅ Company names
- ✅ Locations (Swedish cities)
- ✅ Job descriptions
- ✅ Salary information (when available)
- ✅ Direct application links
- ✅ Remote work flags
- ✅ Skills/requirements

## Search Queries

The connector searches for:
1. **developer** - Software developers, web developers
2. **engineer** - Software engineers, DevOps engineers
3. **designer** - UI/UX designers, graphic designers
4. **manager** - Project managers, product managers
5. **sales** - Sales representatives, account managers
6. **marketing** - Marketing specialists, digital marketers

All searches are for **Sweden** (Sverige).

## Rate Limiting

- **2 seconds** between search queries
- **30 second** HTTP timeout
- Respectful to Jooble API

## Job Transformation

Jooble jobs are transformed to OpenJobs format:

```go
JobPost{
    ID:              "jooble-{job_id}",
    Title:           "Software Developer",
    Company:         "Tech Company AB",
    Description:     "Full job description...",
    Location:        "Stockholm, Sweden",
    Salary:          "50000-70000 SEK/month",
    SalaryCurrency:  "SEK",
    IsRemote:        true/false,
    URL:             "https://...",
    EmploymentType:  "Full-time",
    PostedDate:      time.Time,
    Requirements:    []string{"Python", "React", ...},
    Fields: {
        "source": "jooble",
        "jooble_source": "Indeed", // Original source
        ...
    }
}
```

## Integration with OpenJobs

This plugin is called by the OpenJobs scheduler:

```go
// internal/scheduler/scheduler.go
pluginURLs := map[string]string{
    "jooble": os.Getenv("PLUGIN_JOOBLE_URL"),
}

pluginNames := map[string]string{
    "jooble": "Jooble",
}
```

Default URL: `http://localhost:8088`

## Troubleshooting

### No API Key
If `JOOBLE_API_KEY` is not set, the connector returns demo data (2 sample jobs).

### API Errors
Check the logs for specific error messages from Jooble API.

### No Jobs Returned
- Verify API key is correct
- Check if Jooble API is accessible
- Try demo mode to test connector logic

## Advantages

1. **Free** - No cost for API access
2. **Aggregated** - Jobs from multiple sources
3. **Swedish coverage** - Good local job market coverage
4. **Easy integration** - Simple REST API
5. **Reliable** - Established job search engine

## Limitations

1. **Rate limits** - Unknown (need to test)
2. **Data freshness** - Depends on Jooble's aggregation
3. **Duplicate sources** - May overlap with Indeed-Chrome
4. **API key required** - Need to register

## License

Part of OpenJobs project.
