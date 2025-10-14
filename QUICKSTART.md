# OpenJobs Quick Start Guide

Get OpenJobs running in 5 minutes!

## Prerequisites

- Supabase account (free tier is fine)
- Go 1.21+ OR Docker

## 🚀 Quick Setup (3 Steps)

### Step 1: Configure Supabase

1. Go to https://supabase.com/dashboard
2. Create a new project (or use existing)
3. Go to **SQL Editor** and run:
   ```sql
   -- Copy and paste contents from migrations/001_create_job_posts.sql
   ```
4. Go to **Settings** > **API** and copy:
   - Project URL
   - anon/public key

### Step 2: Configure Environment

```bash
# Copy example config
cp .env.example .env

# Edit .env and add YOUR credentials:
nano .env  # or use your preferred editor
```

Update these lines in `.env`:
```env
SUPABASE_URL=https://your-actual-project.supabase.co
SUPABASE_ANON_KEY=your-actual-anon-key-here
```

⚠️ **CRITICAL**: Replace placeholder values with your actual Supabase credentials!

### Step 3: Run

**Option A: Using Go (Development)**
```bash
go build -o openjobs ./cmd/openjobs
./openjobs
```

**Option B: Using Docker (Production)**
```bash
docker build -t openjobs .
docker run -p 8080:8080 \
  -e SUPABASE_URL="your-url" \
  -e SUPABASE_ANON_KEY="your-key" \
  openjobs
```

## ✅ Test It Works

Run the test script:
```bash
./test_local.sh
```

Or manually:
```bash
# Health check
curl http://localhost:8080/health

# Trigger job sync (fetches jobs from Arbetsförmedlingen)
curl -X POST http://localhost:8080/sync/manual

# View jobs
curl http://localhost:8080/jobs
```

## 🎉 Success!

If you see jobs in the response, you're all set! The system will automatically sync new jobs every 6 hours.

## 📊 Monitor Your Data

Check Supabase dashboard:
1. Go to **Table Editor**
2. Select `job_posts` table
3. See your synchronized jobs!

## 🚀 Deploy to Easypanel

1. Push code to GitHub
2. In Easypanel, create new service from Docker
3. Set environment variables:
   - `SUPABASE_URL`
   - `SUPABASE_ANON_KEY`
4. Deploy!

The Docker container includes automatic cron scheduling (syncs every 6 hours).

## 🔧 Troubleshooting

### "SUPABASE_URL is not configured properly"
- You forgot to update `.env` with real credentials
- Solution: Edit `.env` and add your actual Supabase URL

### "Supabase error 401"
- Wrong API key
- Solution: Copy the correct anon key from Supabase dashboard

### "relation job_posts does not exist"
- Migration not run
- Solution: Run the SQL migration in Supabase SQL Editor

### No jobs fetched
- Arbetsförmedlingen API might be down or changed
- Check logs for specific error messages
- Try manual sync again: `curl -X POST http://localhost:8080/sync/manual`

## 📚 Full Documentation

- [Complete Setup Guide](SETUP_GUIDE.md) - Detailed step-by-step instructions
- [README](README.md) - Project overview and architecture
- [Plugin Development](docs/PLUGIN_DEVELOPMENT.md) - Create custom connectors

## 🎯 Next Steps

1. **Monitor logs** - Watch for automatic syncs every 6 hours
2. **Add more connectors** - Integrate additional job sources
3. **Customize** - Adjust sync frequency, add filters, etc.
4. **Scale** - Deploy to production with Easypanel

## 💡 Pro Tips

- **View logs**: `docker logs -f <container-id>` or check Easypanel logs
- **Manual sync**: Useful for testing or when you want immediate updates
- **Database check**: Always verify Supabase tables are created before first run
- **Test locally first**: Much easier to debug locally than in production

## 🆘 Need Help?

- Check the [SETUP_GUIDE.md](SETUP_GUIDE.md) for detailed troubleshooting
- Review logs for specific error messages
- Ensure Supabase credentials are correct
- Verify migration was run successfully

Happy job hunting! 🎊