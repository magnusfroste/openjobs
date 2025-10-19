# ✅ Indeed Sweden Scraper - EXPERIMENTAL READY!

## ⚠️ IMPORTANT: This is an EXPERIMENTAL connector

**Purpose:** Learn and test web scraping as a method for job aggregation

**Status:** Educational/Testing only - NOT recommended for production

---

## 🎉 What's Been Built

### Files Created:
1. ✅ `/connectors/indeed-scraper/connector.go` - Scraping logic (400+ lines)
2. ✅ `/connectors/indeed-scraper/README.md` - Comprehensive documentation
3. ✅ `/connectors/indeed-scraper/Dockerfile` - Docker container
4. ✅ `/cmd/plugin-indeed-scraper/main.go` - Standalone plugin
5. ✅ `go.mod` - Updated with Colly dependency

---

## 🚀 Quick Start

### Step 1: Install Dependencies

```bash
cd /Users/mafr/Code/OpenJobs

# Download Colly (web scraping library)
go get -u github.com/gocolly/colly/v2

# Download all dependencies
go mod download
```

### Step 2: Check robots.txt (IMPORTANT!)

```bash
# Check if scraping is allowed
curl https://se.indeed.com/robots.txt
```

**Look for:**
- `User-agent: *`
- `Disallow:` directives
- `Crawl-delay:` settings

**If robots.txt disallows scraping, DO NOT USE in production!**

### Step 3: Run the Scraper

```bash
# Test run
go run cmd/plugin-indeed-scraper/main.go
```

**Expected output:**
```
⚠️  ========================================
⚠️  EXPERIMENTAL: Indeed Scraper Plugin
⚠️  Web scraping - Use with caution!
⚠️  Check robots.txt before production use
⚠️  ========================================

🚀 Indeed Scraper Plugin starting on port 8086...
📍 Endpoints:
   GET  /health - Health check
   POST /sync   - Trigger scraping sync
   GET  /jobs   - List scraped jobs

⏱️  Rate limit: 2 seconds between requests
🔍 Queries: developer, engineer, designer, manager, sales
📄 Pages per query: 3 (30 jobs)
📊 Expected: ~100-120 unique jobs per sync
```

---

## 🧪 Testing

### 1. Health Check
```bash
curl http://localhost:8086/health
```

**Expected:**
```json
{
  "status": "healthy",
  "connector": "indeed-scraper",
  "country": "se",
  "method": "web_scraping",
  "experimental": true,
  "warning": "Check robots.txt before production use"
}
```

### 2. Trigger Scraping
```bash
curl -X POST http://localhost:8086/sync
```

**Expected output:**
```
🔄 Scraping sync triggered via HTTP
⚠️  This may take 2-3 minutes due to rate limiting...
🔍 Scraping Indeed for: 'developer'
🔍 Scraping Indeed for: 'engineer'
🔍 Scraping Indeed for: 'designer'
🔍 Scraping Indeed for: 'manager'
🔍 Scraping Indeed for: 'sales'
📊 Scraped 115 unique jobs from Indeed
✅ Stored job: Senior Developer at Tech Company AB (Stockholm, Sweden)
✅ Stored job: Software Engineer at Startup Inc (Göteborg, Sweden)
...
🎉 Indeed scraping sync complete! Fetched: 115, Inserted: 115, Duplicates: 0
```

### 3. View Scraped Jobs
```bash
curl http://localhost:8086/jobs | jq
```

---

## 📊 How It Works

### Scraping Process:

```
1. Build search URL
   ↓
2. Send HTTP request (with realistic User-Agent)
   ↓
3. Parse HTML response
   ↓
4. Extract data using CSS selectors:
   - Job title: h2.jobTitle span[title]
   - Company: span.companyName
   - Location: div.companyLocation
   - Snippet: div.job-snippet
   ↓
5. Transform to JobPost format
   ↓
6. Store in database
   ↓
7. Wait 2 seconds (rate limit)
   ↓
8. Repeat for next page/query
```

### Search Queries:
- `developer` (tech jobs)
- `engineer` (engineering)
- `designer` (design)
- `manager` (management)
- `sales` (sales roles)

### Pages Per Query:
- Page 1 (start=0)
- Page 2 (start=10)
- Page 3 (start=20)

**Total: 5 queries × 3 pages = 15 pages**

---

## ⚠️ Important Limitations

### 1. Legal Gray Area
- Web scraping may violate Indeed's Terms of Service
- Always check robots.txt
- Use only for testing/learning
- NOT recommended for production

### 2. Fragile
- Indeed changes HTML frequently
- Selectors may break
- Requires maintenance
- No guarantees

### 3. Limited Data
- Only snippet (not full description)
- No exact posting dates
- May miss salary info
- Limited to search results

### 4. Slow
- 2 seconds between requests
- ~2-3 minutes per sync
- Can't scale easily

### 5. Risk of Blocking
- Too many requests → CAPTCHA
- IP blocking possible
- Need proxy rotation for scale

---

## 🎯 When to Use This Connector

### ✅ Good For:
- **Learning** - Understand web scraping
- **Experimentation** - Test if scraping works
- **Proof of concept** - Before API partnership
- **Backup** - When API unavailable
- **Small scale** - <100 jobs/day

### ❌ NOT Good For:
- **Production** - Too fragile
- **Large scale** - Too slow
- **Critical systems** - Legal risk
- **Real-time** - Rate limits
- **When API exists** - Use API instead!

---

## 📈 Expected Results

### Per Sync:
- **Time:** 2-3 minutes
- **Requests:** ~15 (5 queries × 3 pages)
- **Jobs fetched:** ~150
- **Unique jobs:** ~100-120
- **Success rate:** 80-90%

### Data Quality:
- ✅ Title: Good
- ✅ Company: Good
- ✅ Location: Good
- ⚠️ Description: Limited (snippet only)
- ⚠️ Salary: Often missing
- ❌ Posted date: Not available

---

## 🔧 Customization

### Add More Queries:
```go
// In connector.go
queries := []string{
    "developer",
    "engineer",
    "designer",
    "manager",
    "sales",
    // Add your own:
    "marketing",
    "data scientist",
    "product manager",
}
```

### Adjust Rate Limit:
```go
rateLimit: 3 * time.Second, // 3 seconds (more respectful)
```

### Scrape More Pages:
```go
// Change from 30 to 50 (5 pages)
for start := 0; start < 50; start += 10 {
    // ...
}
```

---

## 🐛 Troubleshooting

### "No jobs found"
**Possible causes:**
- Indeed changed HTML
- CAPTCHA triggered
- IP blocked
- robots.txt disallows

**Solutions:**
1. Check robots.txt
2. Inspect Indeed's HTML (update selectors)
3. Increase rate limit
4. Use proxy

### "Request timeout"
**Possible causes:**
- Slow network
- Indeed server slow

**Solutions:**
1. Increase timeout (30s → 60s)
2. Check internet connection
3. Try again later

### "Selectors not working"
**Possible causes:**
- Indeed updated HTML

**Solutions:**
1. Visit se.indeed.com/jobs?q=developer
2. Inspect HTML (F12)
3. Find new selectors
4. Update connector.go

---

## 🆚 Comparison: Scraper vs API Connector

| Feature | Scraper | API |
|---------|---------|-----|
| **Setup** | ✅ Easy | ⚠️ Need Publisher ID |
| **Legal** | ⚠️ Gray area | ✅ Official |
| **Reliability** | ❌ Fragile | ✅ Stable |
| **Speed** | ❌ Slow (2s/req) | ✅ Fast |
| **Data** | ⚠️ Limited | ✅ Full |
| **Maintenance** | ❌ High | ✅ Low |
| **Scale** | ❌ Limited | ✅ Unlimited |
| **Cost** | ✅ Free | ⚠️ May cost |

**Verdict:** Scraper is for experimentation, API is for production (but API is discontinued!)

---

## 💡 What We Learned

### ✅ Scraping CAN Work For:
1. Boards without APIs
2. Small-scale testing
3. Proof of concept
4. Backup solution

### ❌ Scraping is HARD Because:
1. HTML changes frequently
2. Legal gray area
3. Rate limiting needed
4. CAPTCHA risk
5. IP blocking
6. High maintenance

### 🎓 Best Practices:
1. **Always check robots.txt**
2. **Generous rate limits** (2+ seconds)
3. **Realistic User-Agent**
4. **Multiple fallback selectors**
5. **Error handling**
6. **Monitoring** (detect when broken)
7. **Caching** (don't re-scrape)

---

## 🚀 Next Steps

### Immediate:
1. ✅ Test scraper locally
2. ✅ Check robots.txt
3. ✅ Verify data quality
4. ✅ Monitor for errors

### Short-term:
1. Add more fallback selectors
2. Improve error handling
3. Add retry logic
4. Monitor selector health

### Long-term:
1. Evaluate if scraping is worth it
2. Consider API partnerships instead
3. Use scraping only for boards without APIs
4. Build scrapers for: The Hub, Academic Work

---

## ⚖️ Legal Disclaimer

**This connector is EXPERIMENTAL and for EDUCATIONAL purposes only.**

**Before production use:**
- ✅ Review Indeed's Terms of Service
- ✅ Check robots.txt
- ✅ Consult legal counsel
- ✅ Consider API partnership
- ✅ Implement monitoring

**Use at your own risk!**

---

## 📚 Resources

**Colly (Go Scraping Library):**
- http://go-colly.org/
- https://github.com/gocolly/colly

**Web Scraping Best Practices:**
- https://www.scrapehero.com/web-scraping-best-practices/
- https://www.zenrows.com/blog/web-scraping-best-practices

**Legal Info:**
- https://blog.apify.com/is-web-scraping-legal/

**robots.txt:**
- https://www.robotstxt.org/

---

## ✅ Success Criteria

**The scraper works if:**
- ✅ Jobs are extracted
- ✅ Data quality is good
- ✅ No CAPTCHAs
- ✅ Rate limiting respected
- ✅ Jobs stored in database
- ✅ No errors

**Expected:**
- Time: 2-3 minutes
- Jobs: 100-120 unique
- Success: >80%
- Errors: <10%

---

## 🎯 Conclusion

**This scraper demonstrates:**
- ✅ Web scraping is technically possible
- ✅ Can get job data without API
- ⚠️ But it's fragile and risky
- ⚠️ APIs are always better when available

**Use this knowledge to:**
1. Build scrapers for boards without APIs
2. Understand scraping challenges
3. Make informed decisions (API vs scraping)
4. Negotiate API partnerships

**Remember:**
- Scraping = Last resort
- APIs = Always prefer
- Legal = Always check
- Respect = Always rate limit

---

**Built on:** Oct 19, 2025  
**Status:** ⚠️ Experimental - Use with caution  
**Port:** 8086  
**Method:** Web scraping (Colly)  
**Purpose:** Learning & experimentation
