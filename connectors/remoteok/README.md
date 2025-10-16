# RemoteOK Connector

Fetches remote tech jobs from RemoteOK.com - one of the largest remote job boards.

## Features

- **No API key required**: RemoteOK provides a public API
- **Large job pool**: Thousands of remote tech positions
- **Rich metadata**: Tags, company logos, apply URLs
- **Global coverage**: Remote jobs from companies worldwide

## API Details

- **Endpoint**: `https://remoteok.com/api`
- **Rate Limit**: Be respectful, no official limit documented
- **Response Format**: JSON array (first item is metadata, skip it)

## Job Fields

RemoteOK provides:
- `position`: Job title
- `company`: Company name
- `location`: Location (usually "Remote" or specific timezone)
- `description`: Full job description
- `tags`: Array of skills/technologies
- `date`: Posted date
- `slug`: URL-friendly identifier
- `company_logo`: Logo URL
- `apply_url`: Direct application link

## Data Transformation

RemoteOK → OpenJobs mapping:
- `id` → `remoteok-{id}`
- `position` → `title`
- `company` → `company`
- `description` → `description`
- `location` → `location` (always includes "Remote")
- `tags` → `requirements` (skills)
- All jobs marked as `is_remote: true`

## Usage

### Standalone Microservice (Default)
RemoteOK runs as an independent microservice on port 8084.

**Build and run:**
```bash
# Build
docker build -f connectors/remoteok/Dockerfile -t plugin-remoteok .

# Run
docker run -p 8084:8084 -e DATABASE_URL=$DATABASE_URL plugin-remoteok
```

**Or use docker-compose:**
```bash
docker-compose -f docker-compose.plugins.yml up -d plugin-remoteok
```

**API Endpoints:**
- `GET /health` - Health check
- `POST /sync` - Trigger job sync
- `GET /jobs` - Fetch latest jobs (without storing)

**Trigger sync:**
```bash
curl -X POST http://localhost:8084/sync
```

## Example Job

```json
{
  "id": "remoteok-123456",
  "title": "Senior Full Stack Developer",
  "company": "TechCorp",
  "description": "We're looking for...",
  "location": "Remote",
  "employment_type": "Full-time",
  "experience_level": "Mid-level",
  "requirements": ["React", "Node.js", "TypeScript"],
  "benefits": ["Remote work"],
  "fields": {
    "source": "remoteok",
    "source_url": "https://remoteok.com/remote-jobs/123456-senior-full-stack-developer",
    "tags": ["React", "Node.js", "TypeScript"],
    "company_logo": "https://...",
    "apply_url": "https://..."
  }
}
```

## Notes

- RemoteOK doesn't provide salary information
- All jobs are remote by definition
- Tags are used as skill requirements
- Jobs expire after 2 months by default
