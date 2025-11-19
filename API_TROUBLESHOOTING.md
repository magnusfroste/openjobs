# API Troubleshooting Guide

## Common Errors When Posting Jobs

### Error: "Invalid JSON" (400 Bad Request)

**Symptom:**
```json
{
  "success": false,
  "message": "Invalid JSON"
}
```

**Common Causes:**

#### 1. Requirements/Benefits as String Instead of Array

❌ **Wrong:**
```json
{
  "title": "Developer",
  "company": "Acme",
  "description": "...",
  "requirements": "5 years experience\nTypeScript\nReact"  // ❌ String!
}
```

✅ **Correct:**
```json
{
  "title": "Developer",
  "company": "Acme",
  "description": "...",
  "requirements": [  // ✅ Array!
    "5 years experience",
    "TypeScript",
    "React"
  ]
}
```

#### 2. Empty Arrays

❌ **Wrong:**
```json
{
  "requirements": [],  // ❌ Empty array
  "benefits": []       // ❌ Empty array
}
```

✅ **Correct:**
```json
{
  // ✅ Just omit empty fields
}
```

Or with values:
```json
{
  "requirements": ["At least 3 years experience"],  // ✅ Has content
  "benefits": ["Health insurance"]                   // ✅ Has content
}
```

#### 3. Invalid Field Types

❌ **Wrong:**
```json
{
  "salary_min": "50000",      // ❌ String
  "salary_max": "70000",      // ❌ String
  "is_remote": "true"         // ❌ String
}
```

✅ **Correct:**
```json
{
  "salary_min": 50000,        // ✅ Number
  "salary_max": 70000,        // ✅ Number
  "is_remote": true           // ✅ Boolean
}
```

---

### Error: "Missing X-API-Key header" (401 Unauthorized)

**Symptom:**
```json
{
  "success": false,
  "message": "Missing X-API-Key header"
}
```

**Solution:**
Add the `X-API-Key` header to your request:

```bash
curl -X POST https://app-openjobs.katsu6.easypanel.host/jobs \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \  # ✅ Add this!
  -d '{ ... }'
```

---

### Error: "Invalid API key" (401 Unauthorized)

**Symptom:**
```json
{
  "success": false,
  "message": "Invalid API key"
}
```

**Causes:**
1. API key doesn't exist in `companies` table
2. API key is misspelled
3. Using old/revoked API key

**Solution:**
1. Check your API key in OpenJobs_Web dashboard
2. Copy it exactly (including the `opj_` prefix)
3. Register a new company if needed

---

### Error: "Title is required" (400 Bad Request)

**Symptom:**
```json
{
  "success": false,
  "message": "Title is required"
}
```

**Required Fields:**
- `title` (string)
- `company` (string)
- `description` (string)

**Solution:**
Ensure all three required fields are present:

```json
{
  "title": "Senior Developer",      // ✅ Required
  "company": "Acme Inc",            // ✅ Required
  "description": "We are hiring..." // ✅ Required
}
```

---

## Field Reference

### Required Fields
```typescript
{
  title: string        // Job title
  company: string      // Company name
  description: string  // Job description
}
```

### Optional Fields
```typescript
{
  location?: string                    // e.g. "Stockholm, Sweden"
  employment_type?: string             // "full-time", "part-time", "contract", etc.
  salary_min?: number                  // Minimum salary (number, not string!)
  salary_max?: number                  // Maximum salary (number, not string!)
  salary_currency?: string             // "SEK", "USD", "EUR", "GBP"
  is_remote?: boolean                  // true or false (not string!)
  url?: string                         // Application URL
  expires_date?: string                // ISO 8601 format: "2025-12-31T23:59:59Z"
  requirements?: string[]              // Array of strings (not single string!)
  benefits?: string[]                  // Array of strings (not single string!)
  fields?: {                           // Metadata object
    company_id?: string
    tags?: string[]
    [key: string]: any
  }
}
```

---

## Working Examples

### Minimal Job Post
```json
{
  "title": "Developer",
  "company": "Acme Inc",
  "description": "We are looking for a talented developer."
}
```

### Full Job Post
```json
{
  "title": "Senior React Developer",
  "company": "Tech AB",
  "description": "We are looking for an experienced React developer to join our team.",
  "location": "Stockholm, Sweden",
  "employment_type": "full-time",
  "salary_min": 50000,
  "salary_max": 70000,
  "salary_currency": "SEK",
  "is_remote": true,
  "url": "https://techab.com/careers/react-dev",
  "expires_date": "2025-12-31T23:59:59Z",
  "requirements": [
    "5+ years React experience",
    "TypeScript proficiency",
    "Team leadership skills"
  ],
  "benefits": [
    "Health insurance",
    "Remote work",
    "Flexible hours",
    "Professional development budget"
  ],
  "fields": {
    "tags": ["react", "typescript", "remote"]
  }
}
```

---

## Testing Your Request

### 1. Validate JSON
Use a JSON validator before sending:
- https://jsonlint.com/
- `jq` command: `echo '{ ... }' | jq`

### 2. Check Field Types
```bash
# ❌ Wrong - salary as string
"salary_min": "50000"

# ✅ Correct - salary as number
"salary_min": 50000
```

### 3. Test with cURL
```bash
curl -X POST https://app-openjobs.katsu6.easypanel.host/jobs \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d @job.json \
  -v  # Verbose output for debugging
```

---

## Frontend Integration Tips

### React/TypeScript Example

```typescript
// ✅ Correct way to prepare job data
const jobData = {
  title: formData.title,
  company: formData.company,
  description: formData.description,
}

// Only add optional fields if they have values
if (formData.location) jobData.location = formData.location
if (formData.salary_min) jobData.salary_min = parseInt(formData.salary_min)

// Convert textarea to array
if (formData.requirements && formData.requirements.trim()) {
  const reqArray = formData.requirements
    .split('\n')
    .map(r => r.trim())
    .filter(r => r.length > 0)
  
  if (reqArray.length > 0) {
    jobData.requirements = reqArray
  }
}

// POST to API
const response = await fetch(`${API_URL}/jobs`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': apiKey
  },
  body: JSON.stringify(jobData)
})
```

---

## Need Help?

1. **Check API Documentation:** See README.md and SETUP_GUIDE.md
2. **Validate Your JSON:** Use jsonlint.com
3. **Test with cURL:** Start with minimal example
4. **Check Logs:** Look at OpenJobs API logs for detailed errors
5. **Open an Issue:** https://github.com/magnusfroste/openjobs/issues
