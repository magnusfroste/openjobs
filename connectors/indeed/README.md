# Indeed Sweden Connector

Connector for fetching job listings from Indeed.se (Sweden) via the Indeed Publisher API.

## Features

- ✅ Fetches jobs from Indeed Sweden (indeed.se)
- ✅ Multiple search queries for diverse job coverage
- ✅ Incremental sync (only new jobs)
- ✅ Deduplication by job key
- ✅ Remote job detection
- ✅ Keyword extraction for requirements
- ✅ Rate limiting (1 second between requests)

## API Information

**Endpoint:** `http://api.indeed.com/ads/apisearch`  
**Documentation:** https://opensource.indeedeng.io/api-documentation/docs/job-search/  
**Publisher Program:** https://www.indeed.com/publisher

## Setup

### 1. Get Publisher ID

1. Visit: https://www.indeed.com/publisher
2. Sign up for the Indeed Publisher Program (free)
3. Get your Publisher ID from the dashboard
4. Add to environment variables

### 2. Environment Variables

```bash
# Required
INDEED_PUBLISHER_ID=your_publisher_id_here

# Optional (defaults shown)
INDEED_COUNTRY=se  # Sweden
```

### 3. Demo Mode

If no `INDEED_PUBLISHER_ID` is set, the connector runs in **demo mode**:
- Uses `publisher=demo`
- Limited to ~25 results
- Good for testing

## API Parameters

The connector uses these Indeed API parameters:

| Parameter | Value | Description |
|-----------|-------|-------------|
| `publisher` | Your ID | Publisher ID from Indeed |
| `v` | `2` | API version |
| `format` | `json` | Response format |
| `co` | `se` | Country (Sweden) |
| `q` | Various | Search query |
| `limit` | `25` | Results per page |
| `start` | `0, 25, 50...` | Pagination offset |

## Search Queries

The connector runs multiple queries to get diverse jobs:

1. **All jobs** - General search
2. **Developer** - Tech jobs
3. **Engineer** - Engineering roles
4. **Manager** - Management positions
5. **Sales** - Sales roles
6. **Customer service** - Service jobs

Each query fetches up to 100 results (4 pages × 25 results).

## Data Mapping

### Indeed → OpenJobs

| Indeed Field | OpenJobs Field | Notes |
|--------------|----------------|-------|
| `jobkey` | `id` | Prefixed with "indeed-" |
| `jobtitle` | `title` | Job title |
| `company` | `company` | Company name |
| `snippet` | `description` | HTML tags removed |
| `formattedLocation` | `location` | City, State, Country |
| `url` | `url` | Direct application link |
| `date` | `posted_date` | Parsed to timestamp |
| - | `salary_currency` | Set to "SEK" (Sweden) |
| - | `is_remote` | Detected from keywords |

## Remote Detection

Jobs are marked as remote if they contain keywords:
- English: `remote`, `work from home`, `wfh`, `anywhere`
- Swedish: `distans`, `hemarbete`, `hemifrån`

## Keyword Extraction

Extracts tech skills and keywords from title and description:
- Programming languages (Java, Python, JavaScript, etc.)
- Frameworks (React, Angular, Django, etc.)
- Tools (Docker, Kubernetes, AWS, etc.)
- Databases (PostgreSQL, MongoDB, etc.)
- Languages (Swedish, English)

## Rate Limiting

- **1 second delay** between API requests
- Respects Indeed's fair use policy
- ~6 requests per search query (6 queries × 4 pages)
- Total: ~24 requests per sync (~30 seconds)

## Expected Results

**Per sync:**
- Queries: 6
- Pages per query: 4
- Results per page: 25
- **Total: ~600 jobs** (before deduplication)
- **Unique: ~400-500 jobs** (after deduplication)

## Testing

### Test with Demo Mode

```bash
# No API key needed
go run cmd/plugin-indeed/main.go
```

### Test with Real API Key

```bash
export INDEED_PUBLISHER_ID=your_id_here
go run cmd/plugin-indeed/main.go
```

## Troubleshooting

### "INDEED_PUBLISHER_ID not set"

**Solution:** Set environment variable or use demo mode for testing.

### "API error 403: Forbidden"

**Possible causes:**
- Invalid Publisher ID
- Publisher account suspended
- Rate limit exceeded

**Solution:** Check your Publisher ID and account status.

### "No results returned"

**Possible causes:**
- No jobs match the query in Sweden
- API is down
- Rate limit hit

**Solution:** Try different queries or wait before retrying.

## Limitations

### Indeed API Limitations

1. **No salary data** - Indeed API doesn't return salary information
2. **Limited fields** - Only basic job info (title, company, snippet, location)
3. **Snippet only** - Full description not available via API
4. **No employment type** - Full-time/part-time not specified
5. **Rate limits** - Fair use policy applies

### Connector Limitations

1. **Swedish jobs only** - Configured for `co=se`
2. **Fixed queries** - Predefined search terms
3. **25 results per page** - API limitation
4. **100 results per query** - Self-imposed limit

## Future Improvements

- [ ] Add more search queries (industry-specific)
- [ ] Support multiple countries
- [ ] Parse salary from snippet text
- [ ] Detect employment type from snippet
- [ ] Add job category classification
- [ ] Implement smarter deduplication
- [ ] Add job freshness scoring

## API Response Example

```json
{
  "version": 2,
  "query": "developer",
  "location": "Sweden",
  "totalResults": 5432,
  "start": 0,
  "end": 24,
  "pageNumber": 0,
  "results": [
    {
      "jobtitle": "Senior Developer",
      "company": "Tech Company AB",
      "city": "Stockholm",
      "state": "Stockholm",
      "country": "SE",
      "formattedLocation": "Stockholm, Sweden",
      "source": "Tech Company AB",
      "date": "Mon, 19 Oct 2025 10:30:00 GMT",
      "snippet": "We are looking for a <b>Senior Developer</b> with experience in React and Node.js...",
      "url": "https://se.indeed.com/viewjob?jk=abc123",
      "latitude": 59.3293,
      "longitude": 18.0686,
      "jobkey": "abc123",
      "sponsored": false,
      "expired": false,
      "formattedRelativeTime": "2 days ago"
    }
  ]
}
```

## Resources

- [Indeed Publisher Program](https://www.indeed.com/publisher)
- [Indeed API Documentation](https://opensource.indeedeng.io/api-documentation/)
- [Indeed Terms of Service](https://www.indeed.com/legal)

## License

Part of OpenJobs project. See main LICENSE file.
