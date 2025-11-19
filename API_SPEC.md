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
| `company` | string | No | none | Filter by company name (exact match) |

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

# Get jobs for a specific company
GET /jobs?company=Tech%20AB&limit=50

# Get your company's jobs (for portal users)
GET /jobs?company=Your%20Company%20Name
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
      "source": "arbetsformedlingen",
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
        "connector": "arbetsformedlingen",
        "original_id": "12345",
        "source_url": "https://arbetsformedlingen.se/..."
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

### POST /jobs

Create a new job posting.

**URL:** `/jobs`

**Method:** `POST`

**Authentication:** Required (`X-API-Key` header)

**Headers:**

| Header | Type | Required | Description |
|--------|------|----------|-------------|
| `Content-Type` | string | Yes | Must be `application/json` |
| `X-API-Key` | string | Yes | Your company API key (get from [registration](https://openjobs-web.vercel.app/register)) |

**Request Body:**

```json
{
  "title": "Senior React Developer",           // Required
  "company": "Your Company",                   // Required
  "description": "We are looking for...",      // Required
  "source": "openjobs-web",                    // Optional (default: openjobs-api)
  "location": "Stockholm, Sweden",             // Optional
  "employment_type": "full-time",              // Optional
  "salary_min": 50000,                         // Optional (number)
  "salary_max": 70000,                         // Optional (number)
  "salary_currency": "SEK",                    // Optional (SEK, USD, EUR, GBP)
  "is_remote": false,                          // Optional (boolean)
  "url": "https://yourcompany.com/apply",      // Optional
  "posted_date": "2025-11-15T10:00:00Z",      // Optional (ISO 8601, defaults to now)
  "expires_date": "2025-12-31T23:59:59Z",     // Optional (ISO 8601)
  "requirements": [                            // Optional (array of strings)
    "5+ years React experience",
    "TypeScript proficiency"
  ],
  "benefits": [                                // Optional (array of strings)
    "Health insurance",
    "Remote work"
  ],
  "fields": {                                  // Optional (metadata object)
    "tags": ["react", "typescript"],
    "ats_job_id": "gh-123456"
  }
}
```

**Source Field Values:**

Use `source` to track where the job came from:
- `openjobs-web` - Posted via OpenJobs Web portal (auto-set)
- `openjobs-api` - Direct API post (default if not specified)
- `greenhouse-api` - From Greenhouse ATS
- `workable-api` - From Workable ATS
- `lever-api` - From Lever ATS
- `custom-integration` - Custom company integration

See [SOURCE_GUIDE.md](SOURCE_GUIDE.md) for detailed guidance.

**Success Response:**

**Code:** `201 Created`

**Content:**

```json
{
  "success": true,
  "data": {
    "id": "web-123e4567-e89b-12d3-a456-426614174000",
    "title": "Senior React Developer",
    "company": "Your Company",
    "description": "We are looking for...",
    "source": "openjobs-api",
    "location": "Stockholm, Sweden",
    "employment_type": "full-time",
    "salary_min": 50000,
    "salary_max": 70000,
    "salary_currency": "SEK",
    "is_remote": false,
    "is_active": true,
    "url": "https://yourcompany.com/apply",
    "posted_date": "2025-11-19T01:00:00Z",
    "requirements": ["5+ years React experience", "TypeScript proficiency"],
    "benefits": ["Health insurance", "Remote work"],
    "fields": {
      "api_posted": true,
      "posted_at": "2025-11-19T01:00:00Z",
      "tags": ["react", "typescript"]
    }
  },
  "message": "Job created successfully"
}
```

**Error Responses:**

**Code:** `400 Bad Request`

```json
{
  "success": false,
  "message": "Title is required"
}
```

**Code:** `401 Unauthorized`

```json
{
  "success": false,
  "message": "Missing X-API-Key header"
}
```

or

```json
{
  "success": false,
  "message": "Invalid API key"
}
```

**Example:**

```bash
curl -X POST https://app-openjobs.katsu6.easypanel.host/jobs \
  -H "Content-Type: application/json" \
  -H "X-API-Key: opj_your_api_key_here" \
  -d '{
    "title": "Senior Developer",
    "company": "Tech AB",
    "description": "We are hiring a senior developer...",
    "source": "greenhouse-api",
    "location": "Stockholm",
    "employment_type": "full-time",
    "salary_min": 50000,
    "salary_max": 70000,
    "salary_currency": "SEK",
    "is_remote": true,
    "requirements": ["React", "TypeScript", "Node.js"],
    "benefits": ["Health insurance", "Remote work"],
    "fields": {
      "ats_system": "greenhouse",
      "ats_job_id": "gh-123456"
    }
  }'
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
| `id` | string | Unique job identifier (format: `{source}-{id}` or `web-{uuid}`) |
| `title` | string | Job title |
| `company` | string | Company name |
| `description` | string | Full job description |
| `source` | string | Primary ingestion source (e.g., `openjobs-web`, `greenhouse-api`, `arbetsformedlingen`) |
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
| `posted_date` | ISO 8601 | When employer posted the job (can be specified, defaults to now) |
| `expires_date` | ISO 8601 | When job posting expires (optional, nullable) |
| `created_at` | ISO 8601 | When OpenJobs created this record (auto-set by database) |
| `updated_at` | ISO 8601 | When job was last modified in OpenJobs (auto-set by database) |
| `requirements` | string[] | Array of required skills/technologies |
| `benefits` | string[] | Array of job benefits |
| `fields` | object | Additional metadata (source-specific, ATS info, tags, etc.) |

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

### Version 2.1 (2025-11-19)
- ✅ Added `POST /jobs` endpoint for job creation
- ✅ Added `source` field for tracking job ingestion channel
- ✅ Added API key authentication via `X-API-Key` header
- ✅ Added support for ATS integrations (Greenhouse, Workable, Lever)
- ✅ Moved `source` from `fields` to top-level field for better analytics

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
