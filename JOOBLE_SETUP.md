# Jooble Plugin Setup Guide

Quick guide to get the Jooble job aggregator plugin running.

## 1. Get API Key

1. Visit https://jooble.org/api/about
2. Fill out the registration form
3. You'll receive your API key via email
4. **Free** - No credit card required

## 2. Add to Environment

### Local Development (.env)
```bash
JOOBLE_API_KEY=your_api_key_here
PORT=8088
```

### Easypanel Deployment
Add environment variable:
```
JOOBLE_API_KEY=your_api_key_here
```

## 3. Deploy

### Option A: Easypanel (Recommended)

**Container Settings:**
- **Name:** openjobs-jooble
- **Port:** 8088
- **Dockerfile Path:** `connectors/jooble/Dockerfile`
- **Build Context:** `/` (project root)

**Environment Variables:**
```bash
JOOBLE_API_KEY=your_key
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=your_supabase_key
PORT=8088
```

**Build Args (if needed):**
```bash
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=your_supabase_key
```

### Option B: Local Docker

```bash
# Build
docker build -t openjobs-jooble -f connectors/jooble/Dockerfile .

# Run
docker run -p 8088:8088 \
  -e JOOBLE_API_KEY=your_key \
  -e SUPABASE_URL=your_url \
  -e SUPABASE_ANON_KEY=your_key \
  openjobs-jooble
```

### Option C: Local Go

```bash
export JOOBLE_API_KEY=your_key
export SUPABASE_URL=your_url
export SUPABASE_ANON_KEY=your_key
PORT=8088 go run cmd/plugin-jooble/main.go
```

## 4. Test

### Health Check
```bash
curl http://localhost:8088/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "jooble-plugin",
  "version": "1.0.0"
}
```

### Trigger Sync
```bash
curl -X POST http://localhost:8088/sync
```

Expected response:
```json
{
  "status": "success",
  "message": "Jooble sync completed successfully"
}
```

### List Jobs
```bash
curl http://localhost:8088/jobs
```

## 5. Add to Scheduler

The scheduler is already configured! Just set the environment variable:

```bash
PLUGIN_JOOBLE_URL=http://jooble:8088
```

Or for local development:
```bash
PLUGIN_JOOBLE_URL=http://localhost:8088
```

## 6. Verify Integration

Check the scheduler logs for:
```
✅ Jooble HTTP sync completed
```

## Expected Results

**Per Sync:**
- 6 search queries (developer, engineer, designer, manager, sales, marketing)
- ~200-400 jobs total
- ~150-300 unique jobs (after deduplication)
- 15-30 seconds sync time

**Daily Volume:**
- Previous: ~1,100-1,300 jobs/day
- With Jooble: ~1,300-1,700 jobs/day
- **+30% increase!**

## Demo Mode

If you don't have an API key yet, the plugin works in demo mode:
- Returns 2 sample jobs
- Good for testing the integration
- Set `JOOBLE_API_KEY` to enable real data

## Troubleshooting

### No jobs returned
- Check if `JOOBLE_API_KEY` is set correctly
- Verify API key is valid (test at https://jooble.org/api/about)
- Check logs for API errors

### API errors
- Rate limit exceeded: Wait and try again
- Invalid API key: Re-register at jooble.org
- Network issues: Check Jooble API status

### Container won't start
- Verify Dockerfile path: `connectors/jooble/Dockerfile`
- Check build context is project root (`/`)
- Ensure environment variables are set

## API Key Registration

**What you'll need:**
- Email address
- Website/app name (can be "OpenJobs")
- Brief description of use case

**Approval time:**
- Usually instant
- Sometimes up to 24 hours

**Limits:**
- Unknown (need to test)
- Likely generous for free tier
- Contact Jooble if you hit limits

## Production Checklist

- [ ] API key registered and received
- [ ] Environment variable set in Easypanel
- [ ] Container deployed and running
- [ ] Health check passing
- [ ] First sync completed successfully
- [ ] Jobs appearing in database
- [ ] Scheduler calling plugin
- [ ] No errors in logs

## Support

**Jooble API:**
- Documentation: https://jooble.org/api/about
- Support: Via their contact form

**OpenJobs:**
- Check logs: `docker logs openjobs-jooble`
- Test endpoints: `/health`, `/sync`, `/jobs`
- Review README: `connectors/jooble/README.md`

## Next Steps

Once Jooble is running:
1. Monitor job volume increase
2. Check data quality
3. Verify no duplicates with Indeed-Chrome
4. Consider adding more Swedish job boards
5. Optimize search queries if needed

---

**Status:** Ready to deploy! 🚀

**Commit:** a45315b
