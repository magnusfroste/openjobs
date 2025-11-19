# Analytics Queries - Job Source Tracking

## Overview

All jobs in OpenJobs have a `source` field in the `fields` JSONB column that tracks where the job came from. This is essential for analytics and understanding your job pipeline.

## Source Values

### Connectors (Automated)
- `arbetsformedlingen` - Swedish Public Employment Service
- `adzuna` - EURES/European jobs via Adzuna API
- `remotive` - Remote-first jobs platform
- `remoteok` - Remote tech jobs
- `indeed` - Indeed job board
- `indeed-scraper` - Indeed scraper fallback
- `indeed-chrome` - Indeed Chrome scraper
- `jooble` - Jooble aggregator
- `offentligajobb` - Swedish public sector jobs

### Manual Posts
- `web-form` - Posted via OpenJobs_Web frontend
- `api` - Posted directly via API (custom integrations)

## SQL Queries

### 1. Jobs by Source (Last 30 Days)

```sql
SELECT 
  fields->>'source' as source,
  COUNT(*) as job_count,
  COUNT(*) * 100.0 / SUM(COUNT(*)) OVER () as percentage
FROM job_posts
WHERE 
  is_active = true 
  AND posted_date >= NOW() - INTERVAL '30 days'
GROUP BY fields->>'source'
ORDER BY job_count DESC;
```

**Example Output:**
```
source              | job_count | percentage
--------------------|-----------|------------
arbetsformedlingen  | 1250      | 62.5%
remoteok            | 450       | 22.5%
web-form            | 200       | 10.0%
adzuna              | 100       | 5.0%
```

---

### 2. Source Performance Over Time

```sql
SELECT 
  DATE(posted_date) as date,
  fields->>'source' as source,
  COUNT(*) as jobs_posted
FROM job_posts
WHERE 
  posted_date >= NOW() - INTERVAL '7 days'
GROUP BY DATE(posted_date), fields->>'source'
ORDER BY date DESC, jobs_posted DESC;
```

---

### 3. Top Companies by Source

```sql
SELECT 
  fields->>'source' as source,
  company,
  COUNT(*) as job_count
FROM job_posts
WHERE 
  is_active = true
  AND posted_date >= NOW() - INTERVAL '30 days'
GROUP BY fields->>'source', company
ORDER BY source, job_count DESC
LIMIT 50;
```

---

### 4. Remote Jobs by Source

```sql
SELECT 
  fields->>'source' as source,
  COUNT(*) as total_jobs,
  SUM(CASE WHEN is_remote = true THEN 1 ELSE 0 END) as remote_jobs,
  ROUND(
    SUM(CASE WHEN is_remote = true THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 
    2
  ) as remote_percentage
FROM job_posts
WHERE is_active = true
GROUP BY fields->>'source'
ORDER BY remote_percentage DESC;
```

**Example Output:**
```
source              | total_jobs | remote_jobs | remote_percentage
--------------------|------------|-------------|------------------
remoteok            | 450        | 450         | 100.00%
remotive            | 320        | 320         | 100.00%
web-form            | 200        | 120         | 60.00%
arbetsformedlingen  | 1250       | 125         | 10.00%
```

---

### 5. Salary Data Quality by Source

```sql
SELECT 
  fields->>'source' as source,
  COUNT(*) as total_jobs,
  SUM(CASE WHEN salary_min IS NOT NULL THEN 1 ELSE 0 END) as jobs_with_salary,
  ROUND(
    SUM(CASE WHEN salary_min IS NOT NULL THEN 1 ELSE 0 END) * 100.0 / COUNT(*),
    2
  ) as salary_data_percentage
FROM job_posts
WHERE is_active = true
GROUP BY fields->>'source'
ORDER BY salary_data_percentage DESC;
```

---

### 6. Average Job Lifespan by Source

```sql
SELECT 
  fields->>'source' as source,
  COUNT(*) as total_jobs,
  ROUND(AVG(EXTRACT(EPOCH FROM (NOW() - posted_date)) / 86400), 1) as avg_days_active
FROM job_posts
WHERE is_active = true
GROUP BY fields->>'source'
ORDER BY avg_days_active DESC;
```

---

### 7. Manual vs Automated Posts

```sql
SELECT 
  CASE 
    WHEN fields->>'source' IN ('web-form', 'api') THEN 'Manual'
    ELSE 'Automated'
  END as post_type,
  COUNT(*) as job_count,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage
FROM job_posts
WHERE 
  is_active = true
  AND posted_date >= NOW() - INTERVAL '30 days'
GROUP BY post_type;
```

---

### 8. Source Reliability (Jobs Still Active)

```sql
SELECT 
  fields->>'source' as source,
  COUNT(*) as total_posted,
  SUM(CASE WHEN is_active = true THEN 1 ELSE 0 END) as still_active,
  ROUND(
    SUM(CASE WHEN is_active = true THEN 1 ELSE 0 END) * 100.0 / COUNT(*),
    2
  ) as active_rate
FROM job_posts
WHERE posted_date >= NOW() - INTERVAL '90 days'
GROUP BY fields->>'source'
ORDER BY active_rate DESC;
```

---

### 9. Jobs with Application URLs by Source

```sql
SELECT 
  fields->>'source' as source,
  COUNT(*) as total_jobs,
  SUM(CASE WHEN url IS NOT NULL AND url != '' THEN 1 ELSE 0 END) as jobs_with_url,
  ROUND(
    SUM(CASE WHEN url IS NOT NULL AND url != '' THEN 1 ELSE 0 END) * 100.0 / COUNT(*),
    2
  ) as url_percentage
FROM job_posts
WHERE is_active = true
GROUP BY fields->>'source'
ORDER BY url_percentage DESC;
```

---

### 10. Source Growth Trend (Week over Week)

```sql
WITH weekly_counts AS (
  SELECT 
    DATE_TRUNC('week', posted_date) as week,
    fields->>'source' as source,
    COUNT(*) as job_count
  FROM job_posts
  WHERE posted_date >= NOW() - INTERVAL '8 weeks'
  GROUP BY DATE_TRUNC('week', posted_date), fields->>'source'
)
SELECT 
  week,
  source,
  job_count,
  LAG(job_count) OVER (PARTITION BY source ORDER BY week) as previous_week,
  job_count - LAG(job_count) OVER (PARTITION BY source ORDER BY week) as growth
FROM weekly_counts
ORDER BY week DESC, job_count DESC;
```

---

## API Endpoint for Analytics

### Get Jobs by Source

```bash
# Via Supabase REST API
curl "https://your-project.supabase.co/rest/v1/job_posts?select=fields->>source,count&is_active=eq.true" \
  -H "apikey: your-anon-key"
```

### Custom Analytics Endpoint (Future)

```bash
# Proposed endpoint
GET /analytics/sources

Response:
{
  "success": true,
  "data": {
    "sources": [
      {
        "name": "arbetsformedlingen",
        "count": 1250,
        "percentage": 62.5,
        "remote_percentage": 10.0,
        "avg_salary": 45000
      },
      ...
    ],
    "total_jobs": 2000,
    "manual_posts": 200,
    "automated_posts": 1800
  }
}
```

---

## Dashboard Metrics

### Key Metrics to Track

1. **Source Distribution**
   - Pie chart of jobs by source
   - Trend over time

2. **Source Quality**
   - Jobs with complete data (salary, URL, requirements)
   - Active job rate
   - Average time to fill

3. **Source Performance**
   - Application conversion rate (if tracked)
   - Jobs per day by source
   - Growth rate

4. **Manual vs Automated**
   - Ratio of manual to automated posts
   - Quality comparison
   - Cost per job (for paid sources)

---

## Grafana Dashboard Example

```json
{
  "dashboard": {
    "title": "OpenJobs Source Analytics",
    "panels": [
      {
        "title": "Jobs by Source (Last 30 Days)",
        "type": "piechart",
        "targets": [
          {
            "rawSql": "SELECT fields->>'source' as metric, COUNT(*) as value FROM job_posts WHERE is_active = true AND posted_date >= NOW() - INTERVAL '30 days' GROUP BY fields->>'source'"
          }
        ]
      },
      {
        "title": "Source Trend",
        "type": "graph",
        "targets": [
          {
            "rawSql": "SELECT posted_date as time, fields->>'source' as metric, COUNT(*) as value FROM job_posts WHERE posted_date >= NOW() - INTERVAL '90 days' GROUP BY posted_date, fields->>'source' ORDER BY posted_date"
          }
        ]
      }
    ]
  }
}
```

---

## Notes

- All queries assume `is_active = true` for current jobs
- Use `posted_date` for time-based filtering
- `fields` is a JSONB column, use `->>'source'` to extract as text
- Consider indexing: `CREATE INDEX idx_job_source ON job_posts ((fields->>'source'));`

---

## Future Enhancements

1. **Source-specific metadata**
   - Track API costs per source
   - Monitor rate limits
   - Quality scores

2. **A/B Testing**
   - Compare job performance by source
   - Optimize connector priorities

3. **Alerting**
   - Alert when source stops posting
   - Alert on quality degradation
   - Alert on unusual patterns
