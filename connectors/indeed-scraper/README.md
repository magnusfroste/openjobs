# Indeed Sweden Scraper (EXPERIMENTAL)

⚠️ **EXPERIMENTAL CONNECTOR** - Web scraping-based job aggregation

## ⚠️ Important Legal & Ethical Considerations

### Legal Status
- **Web scraping is a legal gray area**
- Indeed's Terms of Service may prohibit automated scraping
- Always check `robots.txt` before scraping
- This connector is for **educational/experimental purposes**

### Ethical Scraping Practices
✅ **What we do:**
- Respect rate limits (2 seconds between requests)
- Use realistic User-Agent
- Only scrape public data
- Don't overload servers
- Cache results to minimize requests

❌ **What we DON'T do:**
- Bypass CAPTCHAs
- Ignore robots.txt
- Make excessive requests
- Scrape personal data
- Circumvent access controls

### robots.txt Compliance

Check Indeed's robots.txt:
```
https://se.indeed.com/robots.txt
```

**Always respect:**
- Crawl-delay directives
- Disallowed paths
- User-agent restrictions

---

## 🎯 Purpose

This connector demonstrates **web scraping** as an alternative to APIs for job aggregation.

**Use cases:**
1. **Learning** - Understand scraping techniques
2. **Experimentation** - Test if scraping is viable
3. **Backup** - Fallback when APIs are unavailable
4. **Proof of concept** - Evaluate scraping for other boards

---

## 🛠️ How It Works

### Technology Stack
- **Colly** - Go's best web scraping framework
- **CSS Selectors** - Target specific HTML elements
- **Rate Limiting** - 2 seconds between requests
- **User-Agent Spoofing** - Appear as normal browser

### Scraping Process

```
1. Build search URL (e.g., se.indeed.com/jobs?q=developer)
2. Send HTTP request with realistic headers
3. Parse HTML response
4. Extract job data using CSS selectors
5. Transform to JobPost format
6. Store in database
7. Wait 2 seconds (rate limit)
8. Repeat for next page/query
```

### Data Extraction

**CSS Selectors used:**
```css
div.job_seen_beacon              /* Job card container */
h2.jobTitle span[title]          /* Job title */
span.companyName                 /* Company name */
div.companyLocation              /* Location */
div.job-snippet                  /* Description snippet */
div.salary-snippet               /* Salary (if available) */
```

**Fallback selectors:**
- Indeed changes HTML frequently
- Multiple selectors for each field
- Graceful degradation if elements missing

---

## 📊 Expected Results

**Per sync:**
- Queries: 5 (developer, engineer, designer, manager, sales)
- Pages per query: 3 (0, 10, 20)
- Results per page: ~10 jobs
- **Total: ~150 jobs** (before deduplication)
- **Unique: ~100-120 jobs** (after deduplication)

**Limitations:**
- Only first 3 pages per query (Indeed limits)
- No exact posting dates (not shown in search results)
- Limited description (only snippet)
- May miss jobs if HTML changes

---

## 🚀 Setup & Usage

### Prerequisites

```bash
# Install Colly
go get -u github.com/gocolly/colly/v2
```

### Environment Variables

```bash
# No API keys needed!
# Just database credentials
SUPABASE_URL=your_url
SUPABASE_ANON_KEY=your_key
```

### Run the Scraper

```bash
# Test scraping
go run cmd/plugin-indeed-scraper/main.go
```

---

## 🧪 Testing

### 1. Test Single Query

```go
// In connector_test.go
func TestScrapePage(t *testing.T) {
    store := storage.NewJobStore()
    scraper := NewIndeedScraperConnector(store)
    
    jobs, err := scraper.scrapePage("developer", 0)
    
    assert.NoError(t, err)
    assert.Greater(t, len(jobs), 0)
}
```

### 2. Check robots.txt

```bash
curl https://se.indeed.com/robots.txt
```

### 3. Verify Rate Limiting

```bash
# Should see 2-second delays between requests
go run cmd/plugin-indeed-scraper/main.go
```

---

## ⚠️ Challenges & Limitations

### 1. HTML Changes
**Problem:** Indeed changes HTML structure frequently  
**Solution:** Multiple fallback selectors

### 2. CAPTCHAs
**Problem:** Too many requests trigger CAPTCHAs  
**Solution:** Rate limiting (2 seconds), realistic User-Agent

### 3. IP Blocking
**Problem:** Repeated scraping from same IP  
**Solution:** Rotate proxies (not implemented), reduce frequency

### 4. Data Quality
**Problem:** Only snippet available, no full description  
**Solution:** Could scrape individual job pages (slower)

### 5. Legal Risk
**Problem:** Scraping may violate ToS  
**Solution:** Use only for experimentation, prefer APIs

---

## 🔄 Comparison: Scraping vs API

| Feature | Web Scraping | API |
|---------|-------------|-----|
| **Reliability** | ❌ Fragile (HTML changes) | ✅ Stable |
| **Legal** | ⚠️ Gray area | ✅ Official |
| **Speed** | ⚠️ Slow (rate limits) | ✅ Fast |
| **Data Quality** | ⚠️ Limited (snippets) | ✅ Full data |
| **Maintenance** | ❌ High (fix selectors) | ✅ Low |
| **Cost** | ✅ Free | ⚠️ May cost |
| **Setup** | ✅ Easy | ⚠️ Need approval |

**Verdict:** APIs are better when available, scraping is a fallback.

---

## 🎓 What We Learned

### ✅ Scraping is Viable For:
1. **Boards without APIs** (e.g., The Hub, Academic Work)
2. **Proof of concept** - Test before API partnership
3. **Small-scale** - <1000 jobs/day
4. **Backup** - When API is down

### ❌ Scraping is NOT Good For:
1. **Large-scale** - 10,000+ jobs/day
2. **Production** - Too fragile
3. **Real-time** - Too slow
4. **Critical systems** - Legal risk

### 💡 Best Practices:
1. **Always check robots.txt**
2. **Use generous rate limits** (2+ seconds)
3. **Realistic User-Agent**
4. **Multiple fallback selectors**
5. **Error handling** (graceful failures)
6. **Caching** (don't re-scrape)
7. **Monitoring** (detect when broken)

---

## 🔧 Customization

### Change Search Queries

```go
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

### Adjust Rate Limit

```go
rateLimit: 3 * time.Second, // 3 seconds (more respectful)
```

### Scrape More Pages

```go
// Change from 30 to 50 (5 pages)
for start := 0; start < 50; start += 10 {
    // ...
}
```

---

## 🐛 Troubleshooting

### "No jobs found"
**Causes:**
- Indeed changed HTML structure
- CAPTCHA triggered
- IP blocked
- robots.txt disallows

**Solutions:**
- Update CSS selectors
- Increase rate limit
- Use proxy
- Check robots.txt

### "Request timeout"
**Causes:**
- Slow network
- Indeed server slow
- Too many requests

**Solutions:**
- Increase timeout (30s → 60s)
- Reduce concurrent requests
- Add retry logic

### "Invalid selector"
**Causes:**
- Indeed updated HTML
- Typo in selector

**Solutions:**
- Inspect Indeed's HTML
- Update selectors
- Add fallbacks

---

## 📈 Future Improvements

### Short-term:
- [ ] Add more fallback selectors
- [ ] Scrape individual job pages (full description)
- [ ] Parse salary from text
- [ ] Detect employment type
- [ ] Extract posting date (if available)

### Medium-term:
- [ ] Proxy rotation
- [ ] CAPTCHA detection
- [ ] Retry logic with exponential backoff
- [ ] Monitoring/alerting when selectors break
- [ ] A/B test different User-Agents

### Long-term:
- [ ] Headless browser (Playwright/Puppeteer)
- [ ] JavaScript rendering
- [ ] Screenshot capture (for debugging)
- [ ] ML-based element detection
- [ ] Auto-update selectors when HTML changes

---

## 🎯 When to Use This Connector

### ✅ Use When:
- Experimenting with scraping
- No API available
- Small-scale testing
- Learning purposes
- Backup for API downtime

### ❌ Don't Use When:
- API is available
- Production system
- Large-scale (>1000 jobs/day)
- Legal concerns
- Need reliability

---

## 📚 Resources

**Colly Documentation:**
- http://go-colly.org/
- https://github.com/gocolly/colly

**Web Scraping Best Practices:**
- https://www.scrapehero.com/web-scraping-best-practices/
- https://www.zenrows.com/blog/web-scraping-best-practices

**Legal Considerations:**
- https://blog.apify.com/is-web-scraping-legal/
- https://www.eff.org/issues/coders/reverse-engineering-faq

**robots.txt Spec:**
- https://www.robotstxt.org/

---

## ⚖️ Legal Disclaimer

This connector is provided for **educational and experimental purposes only**.

**Before using in production:**
1. ✅ Review Indeed's Terms of Service
2. ✅ Check robots.txt compliance
3. ✅ Consult legal counsel
4. ✅ Consider API partnership instead
5. ✅ Implement proper rate limiting
6. ✅ Monitor for ToS changes

**The authors are not responsible for:**
- Violations of Terms of Service
- IP blocking or bans
- Legal consequences
- Data accuracy issues
- Service disruptions

**Use at your own risk!**

---

## 🎉 Success Criteria

**The scraper is working if:**
- ✅ Jobs are extracted from search results
- ✅ Data quality is acceptable (title, company, location)
- ✅ No CAPTCHAs triggered
- ✅ Rate limiting respected (2s between requests)
- ✅ Jobs stored in database
- ✅ No errors in logs

**Expected performance:**
- Scrape time: ~2-3 minutes (5 queries × 3 pages × 2s)
- Jobs per sync: 100-120 unique
- Success rate: >80%
- Duplicate rate: <20%

---

**Status:** ⚠️ Experimental - Use with caution  
**Maintenance:** High - Selectors may break  
**Recommended:** Only for testing/learning
