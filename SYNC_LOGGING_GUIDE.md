# Sync Logging System - Prevent API Over-Polling

**Critical feature to respect external API rate limits and optimize scheduler frequency**

## 🎯 Purpose

Track connector efficiency to:
1. **Prevent over-polling** external APIs (avoid upsetting providers!)
2. **Optimize scheduler** frequency based on real data
3. **Monitor duplicate rates** to adjust sync intervals
4. **Identify issues** quickly with detailed logs

## 📊 What Gets Logged

Every sync operation records:
- **Connector name** - Which plugin ran
- **Start/End time** - Duration of sync
- **Jobs fetched** - Total from external API
- **Jobs inserted** - New jobs added to database
- **Jobs duplicates** - Already existing jobs (skipped)
- **Status** - success, error, or partial
- **Error message** - If sync failed

## 🗄️ Database Schema

```sql
-- Migration: 003_create_sync_logs.sql
CREATE TABLE sync_logs (
    id UUID PRIMARY KEY,
    connector_name VARCHAR(100),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    jobs_fetched INTEGER,
    jobs_inserted INTEGER,
    jobs_duplicates INTEGER,
    status VARCHAR(50),  -- success, error, partial
    error_message TEXT
);
```

## 📈 Efficiency Calculation

```
Efficiency = (jobs_inserted / jobs_fetched) × 100
```

**Efficiency Ratings:**
- 🟢 **High (>80%)** - Mostly new jobs, keep current schedule
- 🟡 **Medium (50-80%)** - Some duplicates, schedule is good
- 🔴 **Low (<50%)** - Many duplicates, reduce frequency!

## 🔧 Implementation

### 1. Run Migration

```bash
# In Supabase dashboard SQL Editor:
# Run: migrations/003_create_sync_logs.sql
```

### 2. Connector Logging (Example)

```go
func (ac *ArbetsformedlingenConnector) SyncJobs() error {
    startTime := time.Now()
    
    jobs, err := ac.FetchJobs()
    if err != nil {
        // Log failed sync
        ac.store.LogSync(&models.SyncLog{
            ConnectorName: ac.GetID(),
            StartedAt:     startTime,
            CompletedAt:   time.Now(),
            JobsFetched:   0,
            JobsInserted:  0,
            JobsDuplicates: 0,
            Status:        "error",
            ErrorMessage:  err.Error(),
        })
        return err
    }

    stored := 0
    duplicates := 0
    
    for _, job := range jobs {
        existing, _ := ac.store.GetJob(job.ID)
        if existing != nil {
            duplicates++
            continue
        }
        ac.store.CreateJob(&job)
        stored++
    }

    // Log successful sync
    ac.store.LogSync(&models.SyncLog{
        ConnectorName:  ac.GetID(),
        StartedAt:      startTime,
        CompletedAt:    time.Now(),
        JobsFetched:    len(jobs),
        JobsInserted:   stored,
        JobsDuplicates: duplicates,
        Status:         "success",
    })

    return nil
}
```

### 3. API Endpoint

```
GET /sync/logs
```

Returns last 20 sync operations with full details.

### 4. Dashboard View

The dashboard automatically displays:
- Plugin name
- Time ago (15m ago, 2h ago, etc.)
- Fetched count
- Inserted count
- Duplicates count
- Efficiency badge (color-coded)

## 📊 Example Dashboard Data

| Plugin | Time | Fetched | Inserted | Duplicates | Efficiency |
|--------|------|---------|----------|------------|------------|
| **RemoteOK** | 15m ago | 96 | **96** | 0 | 🟢 **100%** |
| **Arbetsförmedlingen** | 30m ago | 25 | **5** | 20 | 🔴 **20%** |
| **EURES** | 45m ago | 20 | **15** | 5 | 🟢 **75%** |
| **Remotive** | 1h ago | 15 | **13** | 2 | 🟢 **87%** |

## 🎯 Scheduler Optimization

### Current Schedule: Every 6 hours

**Based on efficiency, adjust:**

1. **RemoteOK (100% efficiency)**
   - ✅ Keep 6-hour schedule
   - Or increase to 4 hours (more new jobs!)

2. **Arbetsförmedlingen (20% efficiency)**
   - ⚠️ Reduce to 12 or 24 hours
   - 80% duplicates = wasting API calls!

3. **EURES (75% efficiency)**
   - ✅ Keep 6-hour schedule
   - Good balance

4. **Remotive (87% efficiency)**
   - ✅ Keep 6-hour schedule
   - Excellent efficiency

### Recommended Actions

```go
// internal/scheduler/scheduler.go

// High efficiency connectors (>80%)
scheduler.AddJob("remoteok", 4*time.Hour)      // Increase frequency
scheduler.AddJob("remotive", 6*time.Hour)       // Keep current

// Medium efficiency (50-80%)
scheduler.AddJob("eures", 6*time.Hour)          // Keep current

// Low efficiency (<50%)
scheduler.AddJob("arbetsformedlingen", 24*time.Hour)  // Reduce frequency
```

## 🚨 Respecting API Providers

**Why this matters:**
- External APIs have rate limits
- Over-polling can get you blocked
- Wastes resources (theirs and yours)
- Professional courtesy

**Best practices:**
- Monitor efficiency weekly
- Adjust schedules based on data
- Add delays between requests
- Cache results when possible
- Respect robots.txt and terms of service

## 📝 Monitoring

### Check Logs

```bash
# Via API
curl http://localhost:8080/sync/logs | jq .

# Via Dashboard
http://localhost:8080/dashboard
```

### Query Database

```sql
-- Recent syncs
SELECT * FROM sync_logs 
ORDER BY started_at DESC 
LIMIT 10;

-- Efficiency by connector
SELECT 
    connector_name,
    AVG(jobs_inserted::float / NULLIF(jobs_fetched, 0) * 100) as avg_efficiency,
    COUNT(*) as sync_count
FROM sync_logs
WHERE status = 'success'
GROUP BY connector_name;

-- Failed syncs
SELECT * FROM sync_logs 
WHERE status = 'error'
ORDER BY started_at DESC;
```

## 🔄 Next Steps

1. ✅ Run migration `003_create_sync_logs.sql`
2. ✅ Restart OpenJobs service
3. ✅ Trigger manual sync to test
4. ✅ Check dashboard for sync history
5. ⏰ Monitor for 1 week
6. 📊 Adjust scheduler based on efficiency data

## 🎉 Benefits

- **Respect API limits** - No more over-polling
- **Optimize costs** - Fewer unnecessary API calls
- **Better monitoring** - See exactly what's happening
- **Data-driven decisions** - Adjust based on real metrics
- **Professional** - Shows respect for external services

---

**Remember:** The goal is to fetch all new jobs without annoying API providers! 🤝
