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

The connector is automatically registered in the scheduler. No configuration needed!

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
