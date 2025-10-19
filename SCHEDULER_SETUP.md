# OpenJobs Scheduler Setup

## ✅ Current Configuration

### **Centralized Scheduling:**
All plugins are triggered by the OpenJobs main container using:

```bash
CRON_SCHEDULE=0 6 * * *
```

**This means:** All plugins sync **once per day at 6:00 AM**

---

## 🔌 Plugin Architecture

### **HTTP Microservices Mode:**
- Each plugin runs as independent container
- OpenJobs calls them via HTTP
- Configured via environment variables

### **Plugin URLs:**

```bash
# In Easypanel OpenJobs container env:
PLUGIN_ARBETSFORMEDLINGEN_URL=http://arbetsformedlingen:8081
PLUGIN_EURES_URL=http://eures:8082
PLUGIN_REMOTIVE_URL=http://remotive:8083
PLUGIN_REMOTEOK_URL=http://remoteok:8084
PLUGIN_INDEED_CHROME_URL=http://indeed-chrome:8087  # NEW!
```

---

## 📊 Current Plugins

| Plugin | Port | Method | Frequency | Status |
|--------|------|--------|-----------|--------|
| **Arbetsförmedlingen** | 8081 | API | Daily 6 AM | ✅ Active |
| **EURES** | 8082 | API | Daily 6 AM | ✅ Active |
| **Remotive** | 8083 | API | Daily 6 AM | ✅ Active |
| **RemoteOK** | 8084 | API | Daily 6 AM | ✅ Active |
| **indeed-chrome** | 8087 | Scraper | Daily 6 AM | ✅ NEW! |

---

## ⏰ Schedule Explanation

### **CRON_SCHEDULE=0 6 * * ***

```
┌───────────── minute (0)
│ ┌───────────── hour (6)
│ │ ┌───────────── day of month (*)
│ │ │ ┌───────────── month (*)
│ │ │ │ ┌───────────── day of week (*)
│ │ │ │ │
0 6 * * *
```

**Means:** Every day at 6:00 AM

---

## 🎯 Why This Schedule Works

### **For APIs (Arbetsförmedlingen, EURES, Remotive, RemoteOK):**
- ✅ **Once per day is fine** - Jobs don't change hourly
- ✅ **6 AM is good** - Fresh jobs for morning users
- ✅ **APIs are fast** - Complete in seconds
- ✅ **No risk** - Official APIs welcome frequent access

### **For Scrapers (indeed-chrome):**
- ✅ **Once per day is PERFECT** - Reduces detection risk
- ✅ **6 AM is safe** - Low traffic time
- ✅ **Sustainable** - Won't trigger blocks
- ✅ **Ethical** - Respects servers

**Verdict:** One schedule fits all! 🎉

---

## 🔧 How It Works

### **1. Cron Triggers (6 AM):**
```
OpenJobs Container
    ↓
Scheduler reads CRON_SCHEDULE
    ↓
Triggers at 6:00 AM
```

### **2. Calls All Plugins:**
```
POST http://arbetsformedlingen:8081/sync  ✅
POST http://eures:8082/sync              ✅
POST http://remotive:8083/sync           ✅
POST http://remoteok:8084/sync           ✅
POST http://indeed-chrome:8087/sync      ✅ NEW!
```

### **3. Each Plugin:**
```
Receives /sync request
    ↓
Fetches jobs (API or scraping)
    ↓
Stores in shared database
    ↓
Returns success/failure
```

---

## 📝 Code Changes Made

### **File:** `/internal/scheduler/scheduler.go`

**Added indeed-chrome to plugin list:**

```go
pluginURLs := map[string]string{
    "arbetsformedlingen": os.Getenv("PLUGIN_ARBETSFORMEDLINGEN_URL"),
    "eures":              os.Getenv("PLUGIN_EURES_URL"),
    "remotive":           os.Getenv("PLUGIN_REMOTIVE_URL"),
    "remoteok":           os.Getenv("PLUGIN_REMOTEOK_URL"),
    "indeed-chrome":      os.Getenv("PLUGIN_INDEED_CHROME_URL"), // NEW!
}

// Default URLs
if pluginURLs["indeed-chrome"] == "" {
    pluginURLs["indeed-chrome"] = "http://localhost:8087"
}

// Plugin names
pluginNames := map[string]string{
    "arbetsformedlingen": "Arbetsförmedlingen",
    "eures":              "EURES",
    "remotive":           "Remotive",
    "remoteok":           "RemoteOK",
    "indeed-chrome":      "Indeed Chrome", // NEW!
}
```

**That's it!** No other changes needed.

---

## 🚀 Deployment Steps

### **1. Add Environment Variable in Easypanel:**

**OpenJobs container:**
```bash
PLUGIN_INDEED_CHROME_URL=http://indeed-chrome:8087
```

### **2. Deploy indeed-chrome Container:**

**New container in Easypanel:**
- Name: `indeed-chrome`
- Port: `8087`
- Image: Build from `/connectors/indeed-chrome/Dockerfile`
- Env vars:
  ```bash
  SUPABASE_URL=your_url
  SUPABASE_ANON_KEY=your_key
  PORT=8087
  ```

### **3. Rebuild OpenJobs Container:**
```bash
# With updated scheduler.go
docker build -t openjobs .
```

### **4. Done!**
- indeed-chrome will be called at 6 AM
- Along with all other plugins
- No separate scheduling needed

---

## 🎯 Benefits of This Approach

### **✅ Centralized:**
- One schedule for all plugins
- Easy to manage
- Single source of truth

### **✅ Simple:**
- Just add plugin URL
- No complex logic
- Works immediately

### **✅ Flexible:**
- Want different schedule? Change `CRON_SCHEDULE`
- Want to disable plugin? Remove URL
- Want manual sync? Call `/sync/manual`

### **✅ Scalable:**
- Add more plugins easily
- Each plugin independent
- No coordination needed

---

## 📊 Alternative Schedules (If Needed)

### **Every 6 Hours:**
```bash
CRON_SCHEDULE=0 */6 * * *
```
**Use for:** APIs only (not recommended for scrapers)

### **Twice Per Day:**
```bash
CRON_SCHEDULE=0 6,18 * * *
```
**Use for:** 6 AM and 6 PM syncs

### **Every Hour:**
```bash
CRON_SCHEDULE=0 * * * *
```
**Use for:** Never! Too frequent for scrapers

---

## ⚠️ Important Notes

### **For Scrapers:**
- ✅ **Once per day is PERFECT**
- ❌ **Don't increase frequency**
- ⚠️ **Risk of blocks if too frequent**

### **For APIs:**
- ✅ **Once per day is fine**
- ✅ **Could do more if needed**
- ✅ **No risk with official APIs**

### **Current Setup:**
- ✅ **One schedule (6 AM) works for all**
- ✅ **Simple and maintainable**
- ✅ **No exceptions needed**

---

## 🔍 Monitoring

### **Check Sync Logs:**
```bash
# In OpenJobs container
docker logs openjobs | grep "sync"
```

### **Expected Output:**
```
⏰ Cron triggered at: 2025-10-20 06:00:00
🔧 Running manual job sync for all connectors...
✅ Arbetsförmedlingen HTTP sync completed
✅ EURES HTTP sync completed
✅ Remotive HTTP sync completed
✅ RemoteOK HTTP sync completed
✅ Indeed Chrome HTTP sync completed
✅ All scheduled syncs completed
```

---

## 🎉 Summary

**What we did:**
1. ✅ Added indeed-chrome to scheduler
2. ✅ Uses existing `CRON_SCHEDULE=0 6 * * *`
3. ✅ No other plugins changed
4. ✅ Centralized scheduling maintained

**Result:**
- ✅ indeed-chrome syncs daily at 6 AM
- ✅ Same schedule as all other plugins
- ✅ Simple, clean, maintainable

**Perfect solution!** 🎯

---

**Updated:** Oct 19, 2025  
**Schedule:** Daily at 6:00 AM  
**Method:** Centralized cron in OpenJobs container  
**Status:** ✅ Ready to deploy
