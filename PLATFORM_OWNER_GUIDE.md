# OpenJobs Platform Owner Guide

Complete guide for monitoring, managing, and growing your OpenJobs platform.

## 🎯 Platform Overview

**OpenJobs** is a microservices-based job aggregation platform that collects jobs from 6 different sources.

### Current Status
- **6 Active Plugins** (Arbetsförmedlingen, EURES, Remotive, RemoteOK, Indeed Chrome, Jooble)
- **~1,300-1,700 jobs/day** expected volume
- **Microservices architecture** (fault-isolated, independently scalable)
- **Real-time dashboard** with health monitoring

---

## 📊 Dashboard Access

**URL:** `http://your-domain:8080/dashboard`

### What You See:

#### **1. Core Metrics** (Top Cards)
```
┌─────────────────┬─────────────────┬─────────────────┬─────────────────┐
│  Total Jobs     │  Active Plugins │  Remote Jobs    │  Last Sync      │
│  [dynamic]      │  6              │  [%]            │  [time]         │
└─────────────────┴─────────────────┴─────────────────┴─────────────────┘
```

#### **2. Plugin Status** (Health Grid)
```
✅ Arbetsförmedlingen (8081) - 500 jobs - Healthy
✅ EURES (8082)              - 200 jobs - Healthy
✅ Remotive (8083)           - 100 jobs - Healthy
✅ RemoteOK (8084)           - 100 jobs - Healthy
✅ Indeed Chrome (8087)      - 300 jobs - Healthy
✅ Jooble (8088)             - 250 jobs - Healthy
```

#### **3. Sync History** (Table)
Shows last 20 syncs with:
- Plugin name
- Time
- Jobs fetched
- Jobs inserted
- Duplicates
- Efficiency %

---

## 🔍 Platform Metrics API

**NEW:** `GET /platform/metrics`

### Response Structure:
```json
{
  "success": true,
  "data": {
    "total_jobs": 1450,
    "remote_jobs": 350,
    "active_sources": 6,
    
    "growth": {
      "daily_rate": 7.5,
      "jobs_today": 100,
      "jobs_this_week": 700,
      "jobs_this_month": 3000
    },
    
    "sources": [
      {
        "name": "arbetsformedlingen",
        "jobs": 500,
        "percentage": "34.5",
        "efficiency": "95.0",
        "last_sync": "2025-10-19T18:00:00Z",
        "hours_ago": 2
      }
    ],
    
    "trend": [
      {"date": "2025-10-13", "jobs": 850},
      {"date": "2025-10-14", "jobs": 950},
      ...
    ],
    
    "health": {
      "healthy": 6,
      "warning": 0,
      "down": 0,
      "uptime": "100.0%"
    },
    
    "kpis": {
      "avg_jobs_per_source": 242,
      "remote_percentage": "24.1",
      "total_sources": 6,
      "data_quality_score": "95%"
    }
  }
}
```

---

## 📈 Key Performance Indicators (KPIs)

### **1. Job Volume**
- **Target:** 1,500+ jobs/day
- **Current:** ~1,300-1,700/day
- **Growth:** Track daily/weekly/monthly trends

### **2. Source Diversity**
- **Target:** 6 active sources
- **Current:** 6 sources
- **Balance:** No single source > 40% of total

### **3. Data Quality**
- **Target:** 95%+ efficiency
- **Measure:** Jobs inserted / Jobs fetched
- **Monitor:** Duplicate rate < 20%

### **4. System Health**
- **Target:** 99%+ uptime
- **Monitor:** All plugins syncing < 24h
- **Alert:** Any plugin down > 48h

### **5. Remote Job Coverage**
- **Target:** 20-30% remote jobs
- **Current:** ~24%
- **Trend:** Increasing

---

## 🚨 Monitoring & Alerts

### **Health Status Indicators:**

**🟢 Healthy** (Last sync < 24h)
- All systems operational
- No action needed

**🟡 Warning** (Last sync 24-48h)
- Plugin may have issues
- Check logs
- Consider manual sync

**🔴 Critical** (Last sync > 48h)
- Plugin is down
- Immediate action required
- Check container status

### **What to Monitor Daily:**

1. **Total job count** - Should increase daily
2. **Plugin health** - All should be green
3. **Sync history** - No failed syncs
4. **Efficiency** - Should be > 80%

---

## 🔧 Common Actions

### **Trigger Manual Sync**
```bash
# Via Dashboard
Click "Sync All Plugins" button

# Via API
curl -X POST http://your-domain:8080/sync/manual
```

### **Check Plugin Health**
```bash
# Individual plugin
curl http://plugin-name:8088/health

# All plugins
curl http://your-domain:8080/plugins/status
```

### **View Sync Logs**
```bash
curl http://your-domain:8080/sync/logs
```

### **Restart a Plugin**
```bash
# In Easypanel
1. Go to plugin container
2. Click "Restart"
3. Check health after 30s
```

---

## 📊 Source Performance Analysis

### **Expected Performance by Source:**

| Source | Jobs/Day | Efficiency | Reliability | Cost |
|--------|----------|------------|-------------|------|
| **Arbetsförmedlingen** | 500 | 95%+ | ⭐⭐⭐⭐⭐ | Free |
| **EURES** | 200 | 90%+ | ⭐⭐⭐⭐ | Free |
| **Remotive** | 100 | 85%+ | ⭐⭐⭐⭐ | Free |
| **RemoteOK** | 100 | 90%+ | ⭐⭐⭐⭐⭐ | Free |
| **Indeed Chrome** | 300 | 80%+ | ⭐⭐⭐⭐ | Free |
| **Jooble** | 250 | 85%+ | ⭐⭐⭐⭐ | Free |

### **Performance Benchmarks:**

**Good:**
- Efficiency > 80%
- Sync time < 5 minutes
- Duplicate rate < 20%

**Needs Attention:**
- Efficiency < 70%
- Sync time > 10 minutes
- Duplicate rate > 30%

---

## 💰 Platform Economics

### **Cost Analysis:**

**Infrastructure:**
- Main API: ~$5/month (small container)
- 6 Plugins: ~$30/month (6 × $5)
- Database: ~$10/month (Supabase free tier)
- **Total: ~$45/month**

**Cost per Job:**
- 1,500 jobs/day × 30 days = 45,000 jobs/month
- $45 / 45,000 = **$0.001 per job**
- **Extremely cost-effective!**

### **ROI Metrics:**

**If you monetize (e.g., LazyJobs):**
- 1,000 users × $5/month = $5,000/month
- Cost: $45/month
- **Profit: $4,955/month**
- **ROI: 11,000%+** 🚀

---

## 🎯 Growth Strategy

### **Phase 1: Optimize Current Sources** (Now)
- ✅ All 6 sources running
- ✅ Dashboard monitoring
- ✅ Health alerts
- 🎯 Target: 1,500 jobs/day

### **Phase 2: Add Swedish Job Boards** (Next)
- The Hub (Swedish tech jobs)
- Academic Work (Swedish student jobs)
- Blocket Jobb (Swedish classifieds)
- Monster.se (Swedish jobs)
- 🎯 Target: 2,500 jobs/day

### **Phase 3: European Expansion** (Future)
- LinkedIn Jobs API (paid)
- StepStone (European jobs)
- Glassdoor API
- 🎯 Target: 5,000 jobs/day

### **Phase 4: Global Coverage** (Long-term)
- US job boards
- Asian markets
- LATAM markets
- 🎯 Target: 10,000+ jobs/day

---

## 🔔 Alert Thresholds

### **Set Up Alerts For:**

**Critical (Immediate Action):**
- Any plugin down > 48h
- Total jobs decreased > 20%
- All syncs failing
- Database connection lost

**Warning (Check Within 24h):**
- Plugin efficiency < 70%
- Sync time > 10 minutes
- Duplicate rate > 30%
- Plugin down 24-48h

**Info (Monitor):**
- New jobs < expected
- Source imbalance (one source > 50%)
- Slow growth rate

---

## 📱 Integration Options

### **Slack Notifications:**
```javascript
// Webhook when sync completes
POST https://hooks.slack.com/...
{
  "text": "✅ OpenJobs sync complete: 150 new jobs added"
}
```

### **Email Reports:**
```
Daily Summary:
- Jobs added: 1,450
- Active sources: 6/6
- Health: 100%
- Top source: Arbetsförmedlingen (500 jobs)
```

### **Grafana Dashboard:**
```
Connect to /platform/metrics endpoint
Create custom visualizations
Set up alerts
```

---

## 🎨 Dashboard Enhancements (Roadmap)

### **Coming Soon:**
1. **📈 Charts** - Visual trend graphs
2. **🔔 Alerts** - In-dashboard notifications
3. **⚡ Quick Actions** - Per-plugin sync buttons
4. **📥 Export** - CSV/JSON data export
5. **🔍 Search** - Find specific jobs
6. **📊 Reports** - Weekly/monthly summaries

---

## 🚀 Quick Start Checklist

**For New Platform Owners:**

- [ ] Access dashboard at `/dashboard`
- [ ] Verify all 6 plugins are healthy (green)
- [ ] Trigger manual sync to test
- [ ] Check sync history for errors
- [ ] Review source breakdown
- [ ] Set up monitoring alerts
- [ ] Bookmark `/platform/metrics` endpoint
- [ ] Schedule daily health checks
- [ ] Plan growth strategy
- [ ] Document any issues

---

## 📞 Support & Resources

**Documentation:**
- `/CONNECTORS_SUMMARY.md` - Plugin details
- `/JOOBLE_SETUP.md` - Jooble configuration
- `/CLEANUP_DEAD_PLUGINS.md` - Maintenance guide

**API Endpoints:**
- `GET /dashboard` - Visual dashboard
- `GET /platform/metrics` - Platform KPIs
- `GET /analytics` - Analytics data
- `GET /plugins/status` - Plugin health
- `GET /sync/logs` - Sync history
- `POST /sync/manual` - Trigger sync

**Health Checks:**
- Main API: `http://localhost:8080/health`
- Plugins: `http://localhost:808X/health`

---

## 🎯 Success Metrics

**You're doing well if:**
- ✅ All plugins healthy (6/6 green)
- ✅ Daily job growth > 0
- ✅ Efficiency > 80% across all sources
- ✅ No failed syncs in last 7 days
- ✅ Remote job coverage 20-30%
- ✅ Cost per job < $0.01

**You're crushing it if:**
- 🚀 2,000+ jobs/day
- 🚀 10+ active sources
- 🚀 99%+ uptime
- 🚀 Growing 10%+ weekly
- 🚀 Monetizing successfully

---

## 💡 Pro Tips

1. **Monitor daily** - Check dashboard every morning
2. **Sync strategically** - Run syncs during off-peak hours
3. **Balance sources** - Don't rely on one source
4. **Test new sources** - Always test in staging first
5. **Track trends** - Weekly reviews show patterns
6. **Automate alerts** - Don't rely on manual checks
7. **Document changes** - Keep a changelog
8. **Plan capacity** - Scale before you need to

---

**Remember:** OpenJobs is a platform, not just a tool. Treat it like a business asset!

**Your job as platform owner:**
- 📊 Monitor performance
- 🔧 Fix issues quickly
- 📈 Plan for growth
- 💰 Optimize costs
- 🎯 Deliver value

**You've got this!** 🚀
