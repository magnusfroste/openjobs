# Monitoring Queries - Source Validation

## Overview

Since we don't use CHECK constraints on the `source` column (to avoid maintenance overhead), we rely on monitoring queries to detect unusual or potentially incorrect source values.

## Detect Unusual Sources

### 1. Find All Unique Sources

```sql
SELECT 
  source,
  COUNT(*) as job_count,
  MIN(posted_date) as first_seen,
  MAX(posted_date) as last_seen
FROM job_posts
GROUP BY source
ORDER BY job_count DESC;
```

**What to look for:**
- Typos (e.g., `openjobs-wbe` instead of `openjobs-web`)
- Unexpected sources
- One-off sources (count = 1)

---

### 2. Recent Unusual Sources (Last 7 Days)

```sql
WITH known_sources AS (
  SELECT unnest(ARRAY[
    'arbetsformedlingen',
    'remoteok',
    'remotive',
    'adzuna',
    'indeed',
    'jooble',
    'openjobs-web',
    'openjobs-api',
    'greenhouse-api',
    'workable-api',
    'lever-api',
    'custom-integration'
  ]) as source
)
SELECT 
  jp.source,
  COUNT(*) as job_count,
  array_agg(DISTINCT jp.company) as companies
FROM job_posts jp
LEFT JOIN known_sources ks ON jp.source = ks.source
WHERE 
  ks.source IS NULL  -- Not in known sources
  AND jp.posted_date >= NOW() - INTERVAL '7 days'
GROUP BY jp.source
ORDER BY job_count DESC;
```

**Example Output:**
```
source              | job_count | companies
--------------------|-----------|------------------
bamboohr-api        | 15        | {TechCorp, StartupAB}
openjobs-wbe        | 2         | {CompanyX}  ⚠️ TYPO!
custom-ats          | 8         | {BigCompany}
```

---

### 3. Potential Typos

```sql
-- Find sources that are similar to known sources (Levenshtein distance)
-- Requires pg_trgm extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

WITH known_sources AS (
  SELECT unnest(ARRAY[
    'openjobs-web',
    'openjobs-api',
    'greenhouse-api',
    'workable-api',
    'lever-api'
  ]) as known_source
),
actual_sources AS (
  SELECT DISTINCT source 
  FROM job_posts 
  WHERE posted_date >= NOW() - INTERVAL '30 days'
)
SELECT 
  a.source as actual,
  k.known_source as similar_to,
  similarity(a.source, k.known_source) as similarity_score
FROM actual_sources a
CROSS JOIN known_sources k
WHERE 
  a.source != k.known_source
  AND similarity(a.source, k.known_source) > 0.5  -- 50% similar
ORDER BY similarity_score DESC;
```

**Example Output:**
```
actual           | similar_to      | similarity_score
-----------------|-----------------|------------------
openjobs-wbe     | openjobs-web    | 0.85  ⚠️ Likely typo!
greenhouse       | greenhouse-api  | 0.75  ⚠️ Missing -api
workble-api      | workable-api    | 0.90  ⚠️ Typo!
```

---

### 4. Sources by Company

```sql
-- See which companies use which sources
SELECT 
  company,
  array_agg(DISTINCT source ORDER BY source) as sources,
  COUNT(*) as total_jobs
FROM job_posts
WHERE posted_date >= NOW() - INTERVAL '90 days'
GROUP BY company
HAVING COUNT(DISTINCT source) > 1  -- Companies using multiple sources
ORDER BY total_jobs DESC;
```

**Example Output:**
```
company      | sources                                    | total_jobs
-------------|--------------------------------------------|-----------
TechCorp     | {greenhouse-api, openjobs-web}            | 45
StartupAB    | {openjobs-api, openjobs-web, workable-api}| 32
```

---

### 5. Source Consistency Check

```sql
-- Check if sources are used consistently (no random variations)
SELECT 
  source,
  COUNT(DISTINCT company) as company_count,
  array_agg(DISTINCT company) as companies
FROM job_posts
WHERE posted_date >= NOW() - INTERVAL '30 days'
GROUP BY source
HAVING COUNT(*) < 5  -- Sources with few jobs (might be typos)
ORDER BY COUNT(*);
```

---

## Alerting Queries

### Alert: New Source Detected

```sql
-- Run this daily to detect new sources
WITH yesterday_sources AS (
  SELECT DISTINCT source 
  FROM job_posts 
  WHERE posted_date >= NOW() - INTERVAL '2 days'
    AND posted_date < NOW() - INTERVAL '1 day'
),
today_sources AS (
  SELECT DISTINCT source 
  FROM job_posts 
  WHERE posted_date >= NOW() - INTERVAL '1 day'
)
SELECT 
  t.source as new_source,
  COUNT(*) as job_count,
  array_agg(DISTINCT jp.company) as companies
FROM today_sources t
LEFT JOIN yesterday_sources y ON t.source = y.source
JOIN job_posts jp ON jp.source = t.source 
  AND jp.posted_date >= NOW() - INTERVAL '1 day'
WHERE y.source IS NULL  -- Not in yesterday's sources
GROUP BY t.source;
```

**Alert if:** Any rows returned

---

### Alert: Suspicious Source Pattern

```sql
-- Detect sources that might be errors
SELECT 
  source,
  COUNT(*) as job_count,
  array_agg(DISTINCT company) as companies
FROM job_posts
WHERE 
  posted_date >= NOW() - INTERVAL '7 days'
  AND (
    -- Contains spaces (should use hyphens)
    source LIKE '% %'
    OR
    -- Contains uppercase (should be lowercase)
    source != LOWER(source)
    OR
    -- Too short (likely incomplete)
    LENGTH(source) < 3
    OR
    -- Contains special chars (except hyphen)
    source ~ '[^a-z0-9-]'
  )
GROUP BY source;
```

**Alert if:** Any rows returned

---

## Cleanup Queries

### Fix Common Typos

```sql
-- Dry run: see what would be fixed
SELECT 
  source as current,
  CASE 
    WHEN source = 'openjobs-wbe' THEN 'openjobs-web'
    WHEN source = 'greenhouse' THEN 'greenhouse-api'
    WHEN source = 'workble-api' THEN 'workable-api'
    WHEN source = 'web-form' THEN 'openjobs-web'
    ELSE source
  END as corrected,
  COUNT(*) as affected_jobs
FROM job_posts
WHERE source IN ('openjobs-wbe', 'greenhouse', 'workble-api', 'web-form')
GROUP BY source;

-- Apply fixes (run after reviewing dry run)
UPDATE job_posts
SET source = CASE 
  WHEN source = 'openjobs-wbe' THEN 'openjobs-web'
  WHEN source = 'greenhouse' THEN 'greenhouse-api'
  WHEN source = 'workble-api' THEN 'workable-api'
  WHEN source = 'web-form' THEN 'openjobs-web'
  ELSE source
END
WHERE source IN ('openjobs-wbe', 'greenhouse', 'workble-api', 'web-form');
```

---

## Dashboard Metrics

### Source Health Score

```sql
WITH source_stats AS (
  SELECT 
    source,
    COUNT(*) as job_count,
    COUNT(DISTINCT company) as company_count,
    MIN(posted_date) as first_seen,
    MAX(posted_date) as last_seen,
    -- Check if it's a known source
    CASE 
      WHEN source IN (
        'arbetsformedlingen', 'remoteok', 'remotive', 'adzuna', 'indeed', 'jooble',
        'openjobs-web', 'openjobs-api', 
        'greenhouse-api', 'workable-api', 'lever-api', 'custom-integration'
      ) THEN 'known'
      ELSE 'unknown'
    END as status
  FROM job_posts
  WHERE posted_date >= NOW() - INTERVAL '30 days'
  GROUP BY source
)
SELECT 
  status,
  COUNT(*) as source_count,
  SUM(job_count) as total_jobs,
  ROUND(SUM(job_count) * 100.0 / SUM(SUM(job_count)) OVER (), 2) as percentage
FROM source_stats
GROUP BY status;
```

**Example Output:**
```
status  | source_count | total_jobs | percentage
--------|--------------|------------|------------
known   | 12           | 8,950      | 95.50%
unknown | 3            | 420        | 4.50%
```

**Target:** >95% of jobs from known sources

---

## Grafana Dashboard

```json
{
  "panels": [
    {
      "title": "Source Distribution",
      "targets": [{
        "rawSql": "SELECT source, COUNT(*) as value FROM job_posts WHERE posted_date >= NOW() - INTERVAL '30 days' GROUP BY source"
      }]
    },
    {
      "title": "Unknown Sources Alert",
      "targets": [{
        "rawSql": "SELECT COUNT(DISTINCT source) as value FROM job_posts WHERE posted_date >= NOW() - INTERVAL '7 days' AND source NOT IN ('arbetsformedlingen', 'remoteok', 'remotive', 'adzuna', 'indeed', 'jooble', 'openjobs-web', 'openjobs-api', 'greenhouse-api', 'workable-api', 'lever-api', 'custom-integration')"
      }],
      "alert": {
        "conditions": [{"evaluator": {"params": [1], "type": "gt"}}]
      }
    }
  ]
}
```

---

## Automation Script

```bash
#!/bin/bash
# check_sources.sh - Run daily via cron

SUPABASE_URL="your-url"
SUPABASE_KEY="your-key"

# Get unusual sources
unusual=$(curl -s "${SUPABASE_URL}/rest/v1/rpc/get_unusual_sources" \
  -H "apikey: ${SUPABASE_KEY}" \
  -H "Authorization: Bearer ${SUPABASE_KEY}")

# Alert if any found
if [ "$(echo $unusual | jq length)" -gt 0 ]; then
  echo "⚠️  Unusual sources detected:"
  echo "$unusual" | jq -r '.[] | "  - \(.source) (\(.count) jobs)"'
  
  # Send to Slack/email/etc
  curl -X POST "https://hooks.slack.com/..." \
    -d "{\"text\": \"Unusual job sources detected: $unusual\"}"
fi
```

---

## Best Practices

1. **Run monitoring queries weekly**
   - Check for new unusual sources
   - Look for potential typos

2. **Set up alerts**
   - New source detected
   - Suspicious patterns
   - Sudden drop in known sources

3. **Document new sources**
   - Add to SOURCE_GUIDE.md
   - Update knownSources map in code

4. **Review quarterly**
   - Consolidate similar sources
   - Fix historical typos
   - Update documentation

---

## Summary

**No CHECK constraint = More flexibility, but requires monitoring**

✅ **Pros:**
- No migrations for new sources
- Companies can use custom names
- Easy to add ATS integrations

⚠️ **Requires:**
- Regular monitoring
- Alerting on unusual sources
- Documentation updates
- Periodic cleanup

**This approach scales better for a growing platform!**
