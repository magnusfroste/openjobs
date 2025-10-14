# OpenJobs Setup Guide

This guide will help you set up OpenJobs locally and deploy it to Easypanel.

## Prerequisites

- Go 1.21+ (for local development)
- Docker (for containerized deployment)
- A Supabase account and project

## Step 1: Set Up Supabase

### 1.1 Create Supabase Project

1. Go to https://supabase.com and sign in
2. Click "New Project"
3. Fill in project details and click "Create new project"
4. Wait for the project to be provisioned

### 1.2 Run Database Migration

1. In your Supabase dashboard, go to **SQL Editor**
2. Click "New query"
3. Copy the contents of `migrations/001_create_job_posts.sql`
4. Paste into the SQL editor
5. Click "Run" or press Ctrl+Enter
6. Verify tables are created by going to **Table Editor**

### 1.3 Get API Credentials

1. Go to **Project Settings** > **API**
2. Note down:
   - **Project URL** (e.g., `https://xxxxx.supabase.co`)
   - **anon/public key** (starts with `eyJ...`)

## Step 2: Configure Environment

### 2.1 Create .env File

```bash
cp .env.example .env
```

### 2.2 Update .env with Your Credentials

Edit `.env` and replace the placeholder values:

```env
SUPABASE_URL=https://your-actual-project-ref.supabase.co
SUPABASE_ANON_KEY=eyJhbGc...your-actual-anon-key
PORT=8080
```

**IMPORTANT**: Make sure to use your **actual** Supabase credentials, not the placeholder values!

## Step 3: Test Locally

### 3.1 Build and Run

```bash
# Build the application
go build -o openjobs ./cmd/openjobs

# Run the application
./openjobs
```

You should see:
```
✅ Loaded .env file
✅ Supabase configuration validated
   URL: https://xxxxx.supabase.co
   Key: eyJhbGc...
📊 Database initialization check
🚀 Starting job ingestion scheduler (every 6h)
⏰ Running scheduled job sync at 2024-01-15 10:00:00
🔄 Starting Arbetsförmedlingen job sync...
OpenJobs API starting on port 8080
```

### 3.2 Test Manual Sync

In another terminal, trigger a manual job sync:

```bash
curl -X POST http://localhost:8080/sync/manual
```

Expected output:
```json
{
  "success": true,
  "message": "Job synchronization completed successfully"
}
```

### 3.3 Check Jobs Were Created

```bash
curl http://localhost:8080/jobs
```

You should see jobs returned in the response.

### 3.4 Verify in Supabase

1. Go to your Supabase dashboard
2. Navigate to **Table Editor**
3. Select the `job_posts` table
4. You should see job entries with data from Arbetsförmedlingen

## Step 4: Docker Deployment

### 4.1 Build Docker Image

```bash
docker build -t openjobs:latest .
```

### 4.2 Run Docker Container Locally

```bash
docker run -p 8080:8080 \
  -e SUPABASE_URL="https://your-project.supabase.co" \
  -e SUPABASE_ANON_KEY="your-anon-key" \
  openjobs:latest
```

### 4.3 Test the Container

```bash
# Health check
curl http://localhost:8080/health

# Manual sync
curl -X POST http://localhost:8080/sync/manual

# Get jobs
curl http://localhost:8080/jobs
```

## Step 5: Deploy to Easypanel

### 5.1 Prepare Easypanel

1. Log in to your Easypanel dashboard
2. Create a new project or select existing one

### 5.2 Deploy as Docker Container

1. In Easypanel, click **"+ Create Service"**
2. Select **"App"**
3. Choose **"Docker Image"**
4. Configuration:
   - **Service Name**: openjobs
   - **Docker Image**: Use your registry or build from source
   - **Port**: 8080

### 5.3 Set Environment Variables

In Easypanel service settings, add these environment variables:

```
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key-here
PORT=8080
```

### 5.4 Deploy from GitHub (Alternative)

If using GitHub source:

1. Push your code to GitHub
2. In Easypanel, select **"GitHub Repository"**
3. Connect your repository
4. Set build command: `docker build -t openjobs .`
5. Set start command: Docker will use the CMD from Dockerfile
6. Add environment variables as above

### 5.5 Configure Domain (Optional)

1. In Easypanel service settings, go to **"Domains"**
2. Add your custom domain or use Easypanel subdomain
3. Enable SSL/HTTPS

## Step 6: Verify Deployment

### 6.1 Check Health

```bash
curl https://your-domain.com/health
```

Expected response:
```json
{
  "success": true,
  "data": {
    "service": "openjobs",
    "status": "healthy",
    "version": "1.0.0"
  }
}
```

### 6.2 Verify Cron Job

The Docker container includes a cron job that runs every 6 hours. Check logs in Easypanel:

```
✅ All services started!
   API PID: 7
   Cron PID: 8
```

### 6.3 Test Manual Sync

```bash
curl -X POST https://your-domain.com/sync/manual
```

### 6.4 Monitor Jobs

```bash
curl https://your-domain.com/jobs?limit=5
```

## Troubleshooting

### Issue: "SUPABASE_URL is not configured properly"

**Solution**: Make sure you've updated `.env` with your actual Supabase URL, not the placeholder.

### Issue: "Supabase error 401: Unauthorized"

**Solution**: 
1. Verify your `SUPABASE_ANON_KEY` is correct
2. Check that RLS (Row Level Security) policies allow public access in Supabase
3. Go to Supabase > Authentication > Policies and ensure job_posts table has appropriate policies

### Issue: "Supabase error 404: relation job_posts does not exist"

**Solution**: Run the migration script in Supabase SQL Editor (see Step 1.2)

### Issue: No jobs are being fetched

**Solution**:
1. Check Arbetsförmedlingen API is accessible: `curl https://links.api.jobtechdev.se/joblinks?limit=1`
2. Review application logs for errors
3. Try manual sync: `curl -X POST http://localhost:8080/sync/manual`

### Issue: Cron not running in Docker

**Solution**: Check Docker logs for supercronic output. The cron job should appear in logs every 6 hours.

## Monitoring

### View Logs (Easypanel)

In Easypanel dashboard:
1. Go to your service
2. Click "Logs" tab
3. Monitor for:
   - Scheduled sync messages (every 6 hours)
   - Job creation confirmations
   - Any errors

### Check Job Count

```bash
# Via API
curl https://your-domain.com/jobs?limit=1

# In Supabase Dashboard
# Go to Table Editor > job_posts
# Check row count
```

## API Documentation

### Endpoints

- `GET /health` - Health check
- `GET /jobs` - List jobs (supports `?limit=20&offset=0`)
- `GET /jobs/{id}` - Get specific job
- `POST /jobs` - Create job (manual)
- `PUT /jobs/{id}` - Update job
- `DELETE /jobs/{id}` - Delete job
- `POST /sync/manual` - Trigger manual sync
- `GET /plugins` - List registered plugins

### Example: Create a Job

```bash
curl -X POST https://your-domain.com/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Software Engineer",
    "company": "TechCorp",
    "description": "We are hiring!",
    "location": "Stockholm, Sweden",
    "employment_type": "Full-time"
  }'
```

## Next Steps

1. **Add more connectors**: Integrate EURES and other job sources
2. **Set up monitoring**: Use tools like Grafana or Datadog
3. **Enable RLS**: Configure Supabase Row Level Security policies
4. **Add authentication**: Protect write endpoints
5. **Scale**: Increase sync frequency or add more data sources

## Support

For issues and questions:
- GitHub Issues: https://github.com/openjobs/openjobs/issues
- Documentation: See README.md