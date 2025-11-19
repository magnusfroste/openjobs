# Analytics Refresh Guide

## Overview

OpenJobs uses a **materialized view** (`job_analytics`) for fast analytics queries. This view needs to be refreshed periodically to show the latest data.

---

## 🔄 When to Refresh

### **Automatic Refresh (Recommended)**

Refresh after each connector sync:

```go
// In connector sync handler
func (s *Server) SyncConnector() {
    // 1. Run connector sync
    connector.FetchJobs()
    
    // 2. Refresh analytics
    s.jobStore.RefreshAnalytics()
}
```

### **Manual Refresh**

Via SQL:
```sql
REFRESH MATERIALIZED VIEW job_analytics;
```

Via Supabase Dashboard:
```sql
SELECT refresh_job_analytics();
```

---

## 📊 What Gets Cached

The `job_analytics` materialized view contains:

### **Per Source:**
- `total_jobs` - Total jobs from this source
- `active_jobs` - Currently active jobs
- `remote_jobs` - Remote positions
- `avg_min_salary` - Average minimum salary
- `avg_max_salary` - Average maximum salary
- `countries_covered` - Number of unique countries
- `latest_sync` - When data was last synced
- `jobs_with_url` - Jobs with application URLs
- `jobs_with_requirements` - Jobs with requirements
- `jobs_with_benefits` - Jobs with benefits

---

## 🚀 Implementation

### **1. Migration (Already Done)**

Migration `005_update_analytics_view.sql` creates:
- ✅ Materialized view `job_analytics`
- ✅ Function `refresh_job_analytics()`
- ✅ Function `get_job_analytics_summary()`
- ✅ Function `get_job_analytics_by_source()`

### **2. API Endpoint**

`GET /analytics` now returns **real data**:

```json
{
  "success": true,
  "data": {
    "summary": {
      "total_jobs": 8500,
      "sources_count": 6,
      "countries_covered": 15,
      "remote_percentage": 35
    },
    "sources": [
      {
        "source": "arbetsformedlingen",
        "total_jobs": 5000,
        "active_jobs": 4500,
        "remote_jobs": 200,
        "countries_covered": 1,
        "avg_min_salary": 35000,
        "avg_max_salary": 55000
      }
    ],
    "activity": [
      {
        "timestamp": "2025-11-19T02:00:00Z",
        "event": "sync_completed",
        "source": "arbetsformedlingen",
        "details": "Fetched 150 jobs, inserted 45 new jobs"
      }
    ]
  }
}
```

---

## ⚡ Performance

### **Before (Direct Queries)**
```sql
-- Slow! Runs on every /analytics request
SELECT 
  source,
  COUNT(*) as total_jobs,
  AVG(salary_min) as avg_salary
FROM job_posts
WHERE is_active = true
GROUP BY source;
-- ⏱️ 2-5 seconds with 100k jobs
```

### **After (Materialized View)**
```sql
-- Fast! Pre-computed results
SELECT * FROM job_analytics;
-- ⚡ 10-50 milliseconds
```

**Speedup:** 100-500x faster! 🚀

---

## 🔧 Refresh Strategies

### **Strategy 1: After Each Sync (Recommended)**

```go
// internal/scheduler/scheduler.go
func (s *Scheduler) RunSync() {
    // Sync all connectors
    for _, connector := range s.connectors {
        connector.Sync()
    }
    
    // Refresh analytics once after all syncs
    db.Exec("REFRESH MATERIALIZED VIEW job_analytics")
}
```

**Pros:**
- ✅ Always up-to-date
- ✅ Minimal staleness

**Cons:**
- ❌ Refresh takes time (1-2 seconds)

---

### **Strategy 2: Scheduled Refresh**

```bash
# Cron job - refresh every 30 minutes
*/30 * * * * psql $DATABASE_URL -c "REFRESH MATERIALIZED VIEW job_analytics"
```

**Pros:**
- ✅ Predictable load
- ✅ Doesn't slow down syncs

**Cons:**
- ❌ Data can be 30 minutes stale

---

### **Strategy 3: On-Demand Refresh**

```bash
# Manual trigger via API
POST /analytics/refresh
```

```go
func (s *Server) RefreshAnalytics(w http.ResponseWriter, r *http.Request) {
    db.Exec("REFRESH MATERIALIZED VIEW job_analytics")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
```

**Pros:**
- ✅ Full control
- ✅ No automatic overhead

**Cons:**
- ❌ Requires manual action

---

## 📝 Monitoring

### **Check Last Refresh**

```sql
-- Check when view was last refreshed
SELECT 
  source,
  latest_sync,
  NOW() - latest_sync as staleness
FROM job_analytics
ORDER BY staleness DESC;
```

### **Check View Size**

```sql
-- Check materialized view size
SELECT 
  pg_size_pretty(pg_total_relation_size('job_analytics')) as size;
```

### **Check Refresh Performance**

```sql
-- Time the refresh
\timing on
REFRESH MATERIALIZED VIEW job_analytics;
-- Execution time: 1234.567 ms
```

---

## 🐛 Troubleshooting

### **Issue: Analytics showing old data**

**Solution:** Refresh the view
```sql
REFRESH MATERIALIZED VIEW job_analytics;
```

---

### **Issue: Refresh is slow**

**Possible causes:**
1. Too many jobs (100k+)
2. Complex aggregations
3. Missing indexes

**Solutions:**
```sql
-- Add index on source
CREATE INDEX IF NOT EXISTS idx_job_posts_source ON job_posts(source);

-- Add index on is_active
CREATE INDEX IF NOT EXISTS idx_job_posts_active ON job_posts(is_active);

-- Vacuum and analyze
VACUUM ANALYZE job_posts;
```

---

### **Issue: View doesn't exist**

**Solution:** Run migration
```bash
psql $DATABASE_URL < migrations/005_update_analytics_view.sql
```

---

## 🎯 Best Practices

### **1. Refresh After Bulk Operations**
```go
// After importing 1000s of jobs
ImportJobs(jobs)
RefreshAnalytics()
```

### **2. Don't Refresh Too Often**
```go
// ❌ Bad: Refresh after every single job
for job := range jobs {
    InsertJob(job)
    RefreshAnalytics()  // Too expensive!
}

// ✅ Good: Refresh once after batch
for job := range jobs {
    InsertJob(job)
}
RefreshAnalytics()  // Once at the end
```

### **3. Monitor Refresh Time**
```go
start := time.Now()
RefreshAnalytics()
duration := time.Since(start)
if duration > 5*time.Second {
    log.Printf("⚠️  Analytics refresh took %v (consider optimization)", duration)
}
```

---

## 📊 Example Queries

### **Get Summary**
```sql
SELECT * FROM get_job_analytics_summary();
```

### **Get By Source**
```sql
SELECT * FROM get_job_analytics_by_source();
```

### **Top Sources**
```sql
SELECT 
  source,
  total_jobs,
  active_jobs,
  remote_jobs
FROM job_analytics
ORDER BY total_jobs DESC
LIMIT 10;
```

### **Remote Job Leaders**
```sql
SELECT 
  source,
  remote_jobs,
  ROUND((remote_jobs::decimal / total_jobs) * 100, 2) as remote_percentage
FROM job_analytics
WHERE total_jobs > 0
ORDER BY remote_percentage DESC;
```

---

## ✅ Summary

**Materialized View Benefits:**
- 🚀 100-500x faster queries
- 💾 Pre-computed aggregations
- 📊 Complex analytics without load

**Remember to:**
- ✅ Refresh after syncs
- ✅ Monitor refresh time
- ✅ Check for staleness
- ✅ Optimize if slow

**Current Status:**
- ✅ Migration created
- ✅ JobStore method added
- ✅ AnalyticsHandler updated
- ✅ Real data (no more mocks!)
