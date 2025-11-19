# Portal Integration Guide

## Overview

This guide explains how OpenJobs_Web portal integrates with the OpenJobs API to allow companies to post and view their jobs.

## Architecture

```
┌─────────────────┐
│ OpenJobs_Web    │
│ (Frontend)      │
│                 │
│ - Register      │
│ - Login         │
│ - Post Jobs     │
│ - View My Jobs  │
└────────┬────────┘
         │
         │ HTTP/REST
         │
┌────────▼────────┐
│ OpenJobs API    │
│ (Backend)       │
│                 │
│ - POST /jobs    │
│ - GET /jobs     │
│ - API Key Auth  │
└────────┬────────┘
         │
         │
┌────────▼────────┐
│ Supabase        │
│ (Database)      │
│                 │
│ - companies     │
│ - job_posts     │
└─────────────────┘
```

---

## User Flow

### 1. Company Registration

**Frontend (OpenJobs_Web):**
```typescript
// Register company in Supabase
const { data, error } = await supabase
  .from('companies')
  .insert({
    name: companyName,
    email: email,
    api_key: generateAPIKey(), // opj_xxxxx
  })
  .select()
  .single()

// Store API key for user
localStorage.setItem('openjobs_api_key', data.api_key)
localStorage.setItem('company_name', data.name)
```

**Result:**
- Company gets unique API key
- API key stored in browser
- Used for all subsequent API calls

---

### 2. Post a Job

**Frontend (OpenJobs_Web):**
```typescript
const apiKey = localStorage.getItem('openjobs_api_key')
const companyName = localStorage.getItem('company_name')

const response = await fetch('https://app-openjobs.katsu6.easypanel.host/jobs', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': apiKey,
  },
  body: JSON.stringify({
    title: 'Senior Developer',
    company: companyName,  // Use stored company name
    description: '...',
    source: 'openjobs-web', // Track that it came from web portal
    // ... other fields
  })
})

const result = await response.json()
if (result.success) {
  console.log('Job posted:', result.data.id)
}
```

**Backend (OpenJobs API):**
1. Validates `X-API-Key` header
2. Sets `source = 'openjobs-web'` if not provided
3. Creates job in database
4. Returns job with ID

---

### 3. View My Jobs

**Frontend (OpenJobs_Web):**
```typescript
const companyName = localStorage.getItem('company_name')

// Fetch jobs for this company
const response = await fetch(
  `https://app-openjobs.katsu6.easypanel.host/jobs?company=${encodeURIComponent(companyName)}&limit=100`
)

const result = await response.json()
if (result.success) {
  const myJobs = result.data
  // Display in UI
  myJobs.forEach(job => {
    console.log(job.title, job.posted_date, job.is_active)
  })
}
```

**Key Points:**
- ✅ No authentication needed for GET requests
- ✅ Filter by exact company name
- ✅ URL encode company name (handles spaces)
- ✅ Returns all jobs (active and inactive)

---

## Implementation Examples

### React Component: My Jobs Page

```typescript
import { useState, useEffect } from 'react'

function MyJobsPage() {
  const [jobs, setJobs] = useState([])
  const [loading, setLoading] = useState(true)
  
  useEffect(() => {
    fetchMyJobs()
  }, [])
  
  async function fetchMyJobs() {
    const companyName = localStorage.getItem('company_name')
    if (!companyName) {
      console.error('No company name found')
      return
    }
    
    try {
      const response = await fetch(
        `https://app-openjobs.katsu6.easypanel.host/jobs?company=${encodeURIComponent(companyName)}&limit=100`
      )
      const result = await response.json()
      
      if (result.success) {
        setJobs(result.data)
      }
    } catch (error) {
      console.error('Failed to fetch jobs:', error)
    } finally {
      setLoading(false)
    }
  }
  
  if (loading) return <div>Loading...</div>
  
  return (
    <div>
      <h1>My Jobs ({jobs.length})</h1>
      {jobs.map(job => (
        <div key={job.id}>
          <h3>{job.title}</h3>
          <p>{job.description}</p>
          <span>{job.is_active ? '✅ Active' : '❌ Inactive'}</span>
          <span>Posted: {new Date(job.posted_date).toLocaleDateString()}</span>
        </div>
      ))}
    </div>
  )
}
```

---

### Filter Active vs Inactive Jobs

```typescript
// Get only active jobs
const activeJobs = await fetch(
  `${API_URL}/jobs?company=${encodeURIComponent(companyName)}&is_active=true`
)

// Get only inactive jobs
const inactiveJobs = await fetch(
  `${API_URL}/jobs?company=${encodeURIComponent(companyName)}&is_active=false`
)

// Get all jobs (default)
const allJobs = await fetch(
  `${API_URL}/jobs?company=${encodeURIComponent(companyName)}`
)
```

---

### Pagination for Companies with Many Jobs

```typescript
async function fetchAllMyJobs() {
  const companyName = localStorage.getItem('company_name')
  let allJobs = []
  let offset = 0
  const limit = 100
  
  while (true) {
    const response = await fetch(
      `${API_URL}/jobs?company=${encodeURIComponent(companyName)}&limit=${limit}&offset=${offset}`
    )
    const result = await response.json()
    
    if (!result.success || result.data.length === 0) break
    
    allJobs = [...allJobs, ...result.data]
    offset += limit
    
    // Stop if we got less than limit (last page)
    if (result.data.length < limit) break
  }
  
  return allJobs
}
```

---

## Security Considerations

### ✅ What's Secure

1. **API Key for Posting**
   - POST /jobs requires valid API key
   - Only company with key can post jobs
   - API key validated against database

2. **Company Name Validation**
   - When posting, company name must match registered name
   - Prevents posting jobs for other companies

### ⚠️ What to Consider

1. **Reading Jobs is Public**
   - GET /jobs?company=X doesn't require auth
   - Anyone can see any company's jobs
   - This is by design (jobs are public)

2. **Company Name Must Match**
   - Frontend must use exact company name from registration
   - Case-sensitive matching
   - Spaces and special chars must match

### 🔒 Future Enhancements

1. **User-based Auth**
   - Add user authentication
   - Link users to companies
   - Filter jobs by authenticated user's company

2. **Private Jobs**
   - Add `is_public` field
   - Require auth to view private jobs
   - Only show to company members

---

## Common Issues & Solutions

### Issue 1: No Jobs Returned

**Problem:**
```typescript
const jobs = await fetch(`${API_URL}/jobs?company=My Company`)
// Returns empty array
```

**Cause:** Company name doesn't match exactly

**Solution:**
```typescript
// Check exact company name in database
const { data } = await supabase
  .from('companies')
  .select('name')
  .eq('api_key', apiKey)
  .single()

// Use exact name
const jobs = await fetch(
  `${API_URL}/jobs?company=${encodeURIComponent(data.name)}`
)
```

---

### Issue 2: Special Characters in Company Name

**Problem:**
```typescript
// Company name: "Tech & Co."
const jobs = await fetch(`${API_URL}/jobs?company=Tech & Co.`)
// Fails or returns wrong results
```

**Solution:**
```typescript
// Always URL encode
const companyName = "Tech & Co."
const jobs = await fetch(
  `${API_URL}/jobs?company=${encodeURIComponent(companyName)}`
)
// Becomes: ?company=Tech%20%26%20Co.
```

---

### Issue 3: Case Sensitivity

**Problem:**
```typescript
// Posted as: "TechCorp"
// Querying as: "techcorp"
// No results!
```

**Solution:**
```typescript
// Store company name exactly as registered
localStorage.setItem('company_name', data.name) // "TechCorp"

// Always use stored name
const companyName = localStorage.getItem('company_name')
const jobs = await fetch(`${API_URL}/jobs?company=${encodeURIComponent(companyName)}`)
```

---

## Testing

### Test 1: Post and Retrieve

```bash
# 1. Post a job
curl -X POST https://app-openjobs.katsu6.easypanel.host/jobs \
  -H "Content-Type: application/json" \
  -H "X-API-Key: opj_your_key" \
  -d '{
    "title": "Test Job",
    "company": "Test Company",
    "description": "Test description",
    "source": "openjobs-web"
  }'

# 2. Retrieve it
curl "https://app-openjobs.katsu6.easypanel.host/jobs?company=Test%20Company"

# Should return the job you just posted
```

---

### Test 2: Multiple Jobs

```bash
# Post 3 jobs
for i in 1 2 3; do
  curl -X POST https://app-openjobs.katsu6.easypanel.host/jobs \
    -H "Content-Type: application/json" \
    -H "X-API-Key: opj_your_key" \
    -d "{
      \"title\": \"Job $i\",
      \"company\": \"Test Company\",
      \"description\": \"Description $i\"
    }"
done

# Retrieve all
curl "https://app-openjobs.katsu6.easypanel.host/jobs?company=Test%20Company"

# Should return 3 jobs
```

---

## API Endpoints Summary

### For Portal Users

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/jobs` | POST | ✅ API Key | Post a new job |
| `/jobs?company=X` | GET | ❌ None | View your jobs |
| `/jobs/{id}` | GET | ❌ None | View specific job |

### Query Parameters

| Parameter | Example | Purpose |
|-----------|---------|---------|
| `company` | `?company=TechCorp` | Filter by company |
| `is_active` | `?is_active=true` | Filter active/inactive |
| `limit` | `?limit=50` | Number of results |
| `offset` | `?offset=100` | Pagination |

---

## Next Steps

1. **Implement in OpenJobs_Web**
   - Add "My Jobs" page
   - Use company filter
   - Show active/inactive status

2. **Add Job Management**
   - Edit jobs (UPDATE endpoint)
   - Delete jobs (DELETE endpoint)
   - Toggle active status

3. **Improve UX**
   - Real-time updates
   - Job statistics
   - Analytics dashboard

---

## Support

Questions about portal integration?

- **API Spec:** [API_SPEC.md](API_SPEC.md)
- **Source Guide:** [SOURCE_GUIDE.md](SOURCE_GUIDE.md)
- **GitHub:** https://github.com/magnusfroste/openjobs
