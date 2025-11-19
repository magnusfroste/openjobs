# Source Field Guide - For API Users

## Overview

When posting jobs via the OpenJobs API, you should specify a `source` field to help track where jobs come from. This is important for analytics and understanding your job pipeline.

## What is Source?

The `source` field identifies **how the job entered OpenJobs**, not necessarily where it originated. This helps us (and you) understand:
- Which channels bring the most jobs
- Quality metrics per source
- ROI for different posting methods

## Source Values for API Users

### If You're Posting via OpenJobs_Web Portal

**Automatically set to:** `openjobs-web`

You don't need to do anything - the web portal sets this automatically.

```json
{
  "title": "Developer",
  "company": "Your Company",
  "description": "...",
  // source is automatically set to "openjobs-web"
}
```

---

### If You're Posting via Direct API

**Default:** `openjobs-api`

If you don't specify a source, it defaults to `openjobs-api`.

```bash
curl -X POST https://app-openjobs.katsu6.easypanel.host/jobs \
  -H "X-API-Key: your-api-key" \
  -d '{
    "title": "Developer",
    "company": "Your Company",
    "description": "..."
  }'
# source will be "openjobs-api"
```

---

### If You're Integrating from an ATS

**Recommended:** Specify your ATS name

This helps you track which jobs came from your ATS vs other sources.

#### Greenhouse Integration

```json
{
  "title": "Senior Developer",
  "company": "Your Company",
  "description": "...",
  "source": "greenhouse-api",
  "fields": {
    "ats_system": "greenhouse",
    "ats_job_id": "gh-123456",
    "ats_url": "https://boards.greenhouse.io/yourcompany/jobs/123456"
  }
}
```

#### Workable Integration

```json
{
  "title": "Product Manager",
  "company": "Your Company",
  "description": "...",
  "source": "workable-api",
  "fields": {
    "ats_system": "workable",
    "ats_job_id": "wk-789012",
    "ats_url": "https://apply.workable.com/yourcompany/j/789012"
  }
}
```

#### Lever Integration

```json
{
  "title": "Designer",
  "company": "Your Company",
  "description": "...",
  "source": "lever-api",
  "fields": {
    "ats_system": "lever",
    "ats_job_id": "lv-345678",
    "ats_url": "https://jobs.lever.co/yourcompany/345678"
  }
}
```

---

### If You Have a Custom Integration

**Recommended:** Use `custom-integration` or your system name

```json
{
  "title": "Engineer",
  "company": "Your Company",
  "description": "...",
  "source": "custom-integration",
  "fields": {
    "integration_name": "Your Internal System",
    "internal_job_id": "int-999"
  }
}
```

Or use a descriptive name:

```json
{
  "source": "company-intranet",
  "fields": {
    "posted_from": "internal-hr-portal"
  }
}
```

---

## Valid Source Values

### Recommended for API Users

- `openjobs-web` - Posted via OpenJobs Web portal (auto-set)
- `openjobs-api` - Direct API post (default)
- `greenhouse-api` - From Greenhouse ATS
- `workable-api` - From Workable ATS
- `lever-api` - From Lever ATS
- `custom-integration` - Custom company integration

### Reserved for OpenJobs Connectors

These are used by OpenJobs automated connectors. **Don't use these:**

- `arbetsformedlingen` - Swedish employment service
- `remoteok` - RemoteOK connector
- `remotive` - Remotive connector
- `adzuna` - EURES/Adzuna connector
- `indeed` - Indeed connector
- `jooble` - Jooble connector

---

## Examples by Use Case

### Use Case 1: Small Company, Manual Posts

**Scenario:** You post jobs manually via OpenJobs_Web

**Solution:** Nothing to do! Source is automatically `openjobs-web`

---

### Use Case 2: Company with ATS, Manual Sync

**Scenario:** You copy jobs from Greenhouse and paste into OpenJobs_Web

**Solution:** Still use the web portal, source will be `openjobs-web`

**Optional:** Add ATS info in fields:
```json
{
  "fields": {
    "origin_ats": "greenhouse",
    "synced_manually": true
  }
}
```

---

### Use Case 3: Company with ATS, Automated Sync

**Scenario:** You have a script that syncs jobs from Greenhouse to OpenJobs

**Solution:** Set source to your ATS:

```bash
#!/bin/bash
# Sync script

curl -X POST https://app-openjobs.katsu6.easypanel.host/jobs \
  -H "X-API-Key: $OPENJOBS_API_KEY" \
  -d '{
    "title": "'"$JOB_TITLE"'",
    "company": "'"$COMPANY_NAME"'",
    "description": "'"$JOB_DESC"'",
    "source": "greenhouse-api",
    "url": "'"$GREENHOUSE_URL"'",
    "fields": {
      "ats_system": "greenhouse",
      "ats_job_id": "'"$GREENHOUSE_ID"'",
      "synced_at": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'"
    }
  }'
```

---

### Use Case 4: Multi-Channel Posting

**Scenario:** You post the same job to multiple platforms

**Solution:** Use different sources for each:

```javascript
// Post to OpenJobs Web
await fetch('https://app-openjobs.katsu6.easypanel.host/jobs', {
  method: 'POST',
  headers: {
    'X-API-Key': apiKey,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    ...jobData,
    source: 'openjobs-web'
  })
})

// Also post to LinkedIn
await postToLinkedIn(jobData)

// Track in your system
await trackJobPost({
  job_id: jobData.id,
  channels: ['openjobs-web', 'linkedin'],
  posted_at: new Date()
})
```

---

## Analytics Benefits

### Track Your Posting Channels

```sql
-- See which source brings most jobs
SELECT source, COUNT(*) 
FROM job_posts 
WHERE company = 'Your Company'
GROUP BY source;
```

**Example Output:**
```
source           | count
-----------------|------
openjobs-web     | 45
greenhouse-api   | 120
workable-api     | 30
```

### Compare Performance

```sql
-- Which source gets most applications?
SELECT 
  source,
  COUNT(*) as jobs_posted,
  AVG(application_count) as avg_applications
FROM job_posts
WHERE company = 'Your Company'
GROUP BY source;
```

### Cost Analysis

If you're paying for ATS or other services, track ROI:

```sql
-- Jobs per source per month
SELECT 
  source,
  DATE_TRUNC('month', posted_date) as month,
  COUNT(*) as jobs
FROM job_posts
WHERE company = 'Your Company'
GROUP BY source, month
ORDER BY month DESC, jobs DESC;
```

---

## Best Practices

### 1. Be Consistent

Pick a naming scheme and stick to it:

✅ **Good:**
- `greenhouse-api`
- `workable-api`
- `lever-api`

❌ **Bad:**
- `greenhouse`
- `Workable_API`
- `lever-integration`

### 2. Use Lowercase with Hyphens

✅ **Good:** `custom-integration`  
❌ **Bad:** `Custom_Integration`

### 3. Add Metadata in Fields

Don't put everything in source. Use `fields` for details:

```json
{
  "source": "greenhouse-api",
  "fields": {
    "ats_job_id": "gh-123",
    "ats_department": "Engineering",
    "ats_recruiter": "jane@company.com",
    "sync_version": "2.0"
  }
}
```

### 4. Document Your Sources

Keep a list of sources you use:

```markdown
# Our Job Sources

- `openjobs-web` - Manual posts via web
- `greenhouse-api` - Automated sync from Greenhouse
- `employee-referral` - Jobs from referral program
- `custom-integration` - Internal HR system
```

---

## FAQ

### Q: What if I don't specify a source?

**A:** It defaults to `openjobs-api`. This is fine for simple use cases.

### Q: Can I change the source later?

**A:** Not via API currently. Contact support if you need to bulk update sources.

### Q: Should I use the same source for all my jobs?

**A:** No! Use different sources to track different posting methods. This gives you better analytics.

### Q: What if my ATS isn't listed?

**A:** Use a descriptive name like `bamboohr-api` or `custom-ats`. Just be consistent.

### Q: Can I use custom source names?

**A:** Yes, but check with OpenJobs first to ensure it's not reserved. Use descriptive, lowercase names with hyphens.

---

## Support

Questions about source tracking?

- **Email:** support@openjobs.com
- **Docs:** https://github.com/magnusfroste/openjobs
- **Issues:** https://github.com/magnusfroste/openjobs/issues
