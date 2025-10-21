# 🔧 CORS Fix for OpenJobs API

## Problem
The OpenJobs_Web frontend cannot access the API endpoints due to CORS (Cross-Origin Resource Sharing) restrictions.

## Solution
Added CORS middleware to allow cross-origin requests from the React frontend.

## Files Changed

### 1. `/internal/middleware/cors.go` (NEW)
- CORS middleware that adds necessary headers
- Allows all origins (can be restricted in production)
- Handles OPTIONS preflight requests

### 2. `/cmd/openjobs/main.go` (MODIFIED)
- Added middleware import
- Wrapped all API routes with CORS middleware:
  - `/analytics`
  - `/sync/logs`
  - `/plugins/status`
  - `/sync/manual`
  - `/health`
  - `/plugins`
  - `/plugins/register`
  - `/platform/metrics`

## Deployment Steps

### Option 1: Local Testing

```bash
cd /Users/mafr/Code/OpenJobs
go run cmd/openjobs/main.go
```

### Option 2: Rebuild Docker Image

```bash
cd /Users/mafr/Code/OpenJobs
docker build -t openjobs-api .
docker push your-registry/openjobs-api:latest
```

### Option 3: Easypanel Deployment

1. **Commit changes:**
```bash
cd /Users/mafr/Code/OpenJobs
git add internal/middleware/cors.go cmd/openjobs/main.go
git commit -m "Add CORS middleware for frontend access"
git push
```

2. **Redeploy in Easypanel:**
   - Go to your OpenJobs app
   - Click "Rebuild" or trigger deployment
   - Wait for deployment to complete

3. **Verify:**
```bash
curl -I https://app-openjobs.katsu6.easypanel.host/analytics
# Should see: Access-Control-Allow-Origin: *
```

## Testing

After deployment, test the endpoints:

```bash
# Test analytics endpoint
curl -H "Origin: http://localhost:3001" \
     -I https://app-openjobs.katsu6.easypanel.host/analytics

# Should return:
# Access-Control-Allow-Origin: *
# Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
```

## What This Fixes

✅ OpenJobs_Web `/status` page can now fetch data  
✅ All API endpoints accessible from browser  
✅ No more "Failed to fetch" errors  
✅ CORS preflight requests handled correctly  

## Production Considerations

For production, you may want to restrict origins:

```go
// In internal/middleware/cors.go
w.Header().Set("Access-Control-Allow-Origin", "https://openjobs.yourdomain.com")
```

Or use environment variable:

```go
allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
if allowedOrigin == "" {
    allowedOrigin = "*" // Default to allow all
}
w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
```
