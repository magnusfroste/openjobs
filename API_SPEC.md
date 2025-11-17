# OpenJobs API Specification

**Version:** 2.0  
**Base URL:** `https://app-openjobs.katsu6.easypanel.host`  
**Authentication:** None required for read endpoints

---

## 📡 Endpoints

### GET /jobs

Retrieve a list of job postings with optional filtering and pagination.

**URL:** `/jobs`

**Method:** `GET`

**Authentication:** None required

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 20 | Number of jobs to return (max: 500) |
| `offset` | integer | No | 0 | Number of jobs to skip for pagination |
| `is_active` | boolean | No | all | Filter by active status (`true` or `false`) |
| `created_after` | ISO 8601 | No | none | Return jobs created after this timestamp |

**Examples:**

```bash
# Get first 100 active jobs
GET /jobs?is_active=true&limit=100

# Get jobs created in the last 24 hours
GET /jobs?created_after=2025-11-16T15:00:00Z&is_active=true

# Incremental sync - get only new jobs since last check
GET /jobs?created_after=2025-11-17T06:00:00Z&is_active=true&limit=500

# Pagination - get next page
GET /jobs?is_active=true&limit=100&offset=100
```

**Success Response:**

**Code:** `200 OK`

**Content:**

```json
{
  "success": true,
  "data": [
    {
      "id": "af-12345",
      "title": "Senior Developer",
      "company": "Tech AB",
      "description": "Full job description text...",
      "location": "Stockholm",
      "salary": "50000-70000 SEK",
      "salary_min": 50000,
      "salary_max": 70000,
      "salary_currency": "SEK",
      "is_remote": false,
      "is_active": true,
      "url": "https://arbetsformedlingen.se/...",
      "employment_type": "full-time",
      "experience_level": "senior",
      "posted_date": "2025-11-15T10:00:00Z",
      "expires_date": "2025-12-15T10:00:00Z",
      "created_at": "2025-11-17T06:15:23Z",
      "updated_at": "2025-11-17T06:15:23Z",
      "requirements": ["Python", "React", "PostgreSQL"],
      "benefits": ["Remote work", "Health insurance", "Pension"],
      "fields": {
        "source": "arbetsformedlingen",
        "connector": "arbetsformedlingen",
        "original_id": "12345"
      }
    }
  ]
}
```

**Error Response:**

**Code:** `500 Internal Server Error`

**Content:**

```json
{
  "success": false,
  "message": "Failed to retrieve jobs"
}
```

---

### GET /jobs/:id

Retrieve a specific job by ID.

**URL:** `/jobs/:id`

**Method:** `GET`

**Authentication:** None required

**URL Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | Yes | Job ID (e.g., `af-12345`) |

**Example:**

```bash
GET /jobs/af-12345
```

**Success Response:**

**Code:** `200 OK`

**Content:** Same as single job object in `/jobs` response

---

### GET /health

Check API health status.

**URL:** `/health`

**Method:** `GET`

**Authentication:** None required

**Success Response:**

**Code:** `200 OK`

**Content:**

```json
{
  "status": "healthy",
  "timestamp": "2025-11-17T15:30:00Z"
}
```

---

### POST /sync/manual

Trigger a manual sync of all connectors (admin only).

**URL:** `/sync/manual`

**Method:** `POST`

**Authentication:** Required (admin)

---

## 📊 Response Fields

### Job Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique job identifier (format: `{source}-{id}`) |
| `title` | string | Job title |
| `company` | string | Company name |
| `description` | string | Full job description |
| `location` | string | Job location (city, country) |
| `salary` | string | Salary range as text |
| `salary_min` | integer | Minimum salary (nullable) |
| `salary_max` | integer | Maximum salary (nullable) |
| `salary_currency` | string | Currency code (SEK, USD, EUR) |
| `is_remote` | boolean | Whether job is remote |
| `is_active` | boolean | Whether job is currently active |
| `url` | string | Direct application URL |
| `employment_type` | string | Type of employment (full-time, part-time, contract) |
| `experience_level` | string | Required experience level (junior, mid, senior) |
| `posted_date` | ISO 8601 | When employer originally posted the job |
| `expires_date` | ISO 8601 | When job posting expires (nullable) |
| `created_at` | ISO 8601 | When OpenJobs added the job to database |
| `updated_at` | ISO 8601 | When job was last modified in OpenJobs |
| `requirements` | string[] | Array of required skills/technologies |
| `benefits` | string[] | Array of job benefits |
| `fields` | object | Additional metadata (source-specific) |

---

## 🔄 Incremental Sync Guide

For efficient job syncing, use the `created_after` parameter with the `created_at` field:

### First Sync

```bash
# Fetch all active jobs (or jobs from last 7 days)
GET /jobs?is_active=true&limit=500

# Store the current timestamp
last_sync = "2025-11-17T15:30:00Z"
```

### Subsequent Syncs

```bash
# Only fetch jobs created since last sync
GET /jobs?created_after=2025-11-17T15:30:00Z&is_active=true&limit=500

# Update last_sync timestamp
last_sync = "2025-11-17T16:30:00Z"
```

### Why Use `created_at` Instead of `posted_date`?

- **`posted_date`**: When the employer originally posted the job (can be days/weeks old when first fetched)
- **`created_at`**: When OpenJobs added the job to the database (accurate for incremental sync)

**Example:**
```
Job posted by employer: Nov 12 (posted_date)
OpenJobs fetches job: Nov 17 (created_at)
Your last sync: Nov 15

Query with posted_date: ❌ MISSES the job (Nov 12 < Nov 15)
Query with created_at: ✅ GETS the job (Nov 17 > Nov 15)
```

---

## 📈 Rate Limits

Currently no rate limits enforced. Please be respectful:
- Recommended: Max 10 requests per minute
- Use `created_after` for incremental sync instead of fetching all jobs repeatedly

---

## 🐛 Error Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Bad Request (invalid parameters) |
| 404 | Not Found (job ID doesn't exist) |
| 500 | Internal Server Error |

---

## 📝 Changelog

### Version 2.0 (2025-11-17)
- ✅ Added `is_active` filtering
- ✅ Added `created_after` parameter for incremental sync
- ✅ Added `is_active` field to job response
- ✅ Increased max limit to 500 jobs per request

### Version 1.0 (2025-10-15)
- Initial release
- Basic job listing and retrieval

---

## 🔗 Resources

- **API Base URL:** https://app-openjobs.katsu6.easypanel.host
- **GitHub:** https://github.com/magnusfroste/openjobs
- **Dashboard:** https://openjobs-web.froste.eu (coming soon)

---

**Questions?** Open an issue on GitHub or contact us at support@froste.eu
