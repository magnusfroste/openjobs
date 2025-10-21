# Offentliga Jobb Chrome Scraper

Headless Chrome-based scraper for Swedish public sector jobs from Offentliga Jobb.

## Overview

- **Source:** https://www.offentligajobb.se/
- **Focus:** Swedish public sector (government, municipalities, universities, healthcare)
- **Method:** Headless Chrome (chromedp)
- **Port:** 8089
- **Status:** 🚧 Work in Progress - Needs URL/selector updates

## Quick Start

### Local Development

```bash
export SUPABASE_URL=your_url
export SUPABASE_ANON_KEY=your_key
PORT=8089 go run cmd/plugin-offentligajobb/main.go
```

### Docker

```bash
docker build -t openjobs-offentligajobb -f connectors/offentligajobb/Dockerfile .
docker run -p 8089:8089 \
  -e SUPABASE_URL=your_url \
  -e SUPABASE_ANON_KEY=your_key \
  openjobs-offentligajobb
```

## Endpoints

- `GET /health` - Health check
- `POST /sync` - Trigger scraping
- `GET /jobs` - List scraped jobs

## TODO

This is a clone of Indeed-Chrome that needs adaptation:

1. **Update URLs** - Change from Indeed to Offentliga Jobb
2. **Update Selectors** - Find correct CSS selectors for job listings
3. **Test Scraping** - Verify it works with Offentliga Jobb
4. **Adjust Queries** - Public sector specific searches

## Expected Results

- ~200-400 public sector jobs per sync
- Daily updates
- Swedish government/municipality jobs
- Universities, healthcare, agencies

## Integration

Add to scheduler:
```bash
PLUGIN_OFFENTLIGAJOBB_URL=http://offentligajobb:8089
```

## Status

⚠️ **NEEDS WORK** - This is a quick clone that needs URL and selector updates before it will work!
