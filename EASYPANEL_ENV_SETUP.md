# Easypanel Environment Variables Setup

## Overview

OpenJobs uses a microservices architecture with 5 containers:
1. **Core** (port 8080) - API, Dashboard, Scheduler
2. **Plugin: Arbetsförmedlingen** (port 8081)
3. **Plugin: EURES** (port 8082)
4. **Plugin: Remotive** (port 8083)
5. **Plugin: RemoteOK** (port 8084)

## Core Container Environment Variables

**Service Name:** `openjobs` (or `app-openjobs`)

### Required Variables

```bash
# Supabase Connection
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyAgCiAgICAicm9sZSI6ICJhbm9uIiwKICAgICJpc3MiOiAic3VwYWJhc2UtZGVtbyIsCiAgICAiaWF0IjogMTY0MTc2OTIwMCwKICAgICJleHAiOiAxNzk5NTM1NjAwCn0.dc_X5iR_VP_qT0zsiyj_I_OZ2T9FtRU2BBNWN8Bu4GE

# Microservices Mode (CRITICAL!)
USE_HTTP_PLUGINS=true

# Plugin Container URLs (use Easypanel internal network names)
PLUGIN_ARBETSFORMEDLINGEN_URL=http://plugin-arbetsformedlingen:8081
PLUGIN_EURES_URL=http://plugin-eures:8082
PLUGIN_REMOTIVE_URL=http://plugin-remotive:8083
PLUGIN_REMOTEOK_URL=http://plugin-remoteok:8084

# Port
PORT=8080
```

### Important Notes

- **USE_HTTP_PLUGINS=true** - This tells the core to call plugin containers instead of running connectors locally
- **Plugin URLs** - Use Easypanel's internal Docker network names (not external URLs)
- The format is: `http://<easypanel-service-name>:<port>`

## Plugin Container Environment Variables

All plugin containers need the same Supabase credentials:

### Plugin: Arbetsförmedlingen (port 8081)

```bash
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyAgCiAgICAicm9sZSI6ICJhbm9uIiwKICAgICJpc3MiOiAic3VwYWJhc2UtZGVtbyIsCiAgICAiaWF0IjogMTY0MTc2OTIwMCwKICAgICJleHAiOiAxNzk5NTM1NjAwCn0.dc_X5iR_VP_qT0zsiyj_I_OZ2T9FtRU2BBNWN8Bu4GE
PORT=8081
```

### Plugin: EURES (port 8082)

```bash
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyAgCiAgICAicm9sZSI6ICJhbm9uIiwKICAgICJpc3MiOiAic3VwYWJhc2UtZGVtbyIsCiAgICAiaWF0IjogMTY0MTc2OTIwMCwKICAgICJleHAiOiAxNzk5NTM1NjAwCn0.dc_X5iR_VP_qT0zsiyj_I_OZ2T9FtRU2BBNWN8Bu4GE
PORT=8082
ADZUNA_APP_ID=your-adzuna-id
ADZUNA_APP_KEY=your-adzuna-key
```

### Plugin: Remotive (port 8083)

```bash
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyAgCiAgICAicm9sZSI6ICJhbm9uIiwKICAgICJpc3MiOiAic3VwYWJhc2UtZGVtbyIsCiAgICAiaWF0IjogMTY0MTc2OTIwMCwKICAgICJleHAiOiAxNzk5NTM1NjAwCn0.dc_X5iR_VP_qT0zsiyj_I_OZ2T9FtRU2BBNWN8Bu4GE
PORT=8083
```

### Plugin: RemoteOK (port 8084) ✅ Already Correct

```bash
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyAgCiAgICAicm9sZSI6ICJhbm9uIiwKICAgICJpc3MiOiAic3VwYWJhc2UtZGVtbyIsCiAgICAiaWF0IjogMTY0MTc2OTIwMCwKICAgICJleHAiOiAxNzk5NTM1NjAwCn0.dc_X5iR_VP_qT0zsiyj_I_OZ2T9FtRU2BBNWN8Bu4GE
PORT=8084
```

## Easypanel Service Names

Make sure your Easypanel services are named correctly so the internal network URLs work:

| Service | Expected Name | Port |
|---------|--------------|------|
| Core | `openjobs` or `app-openjobs` | 8080 |
| Arbetsförmedlingen | `plugin-arbetsformedlingen` | 8081 |
| EURES | `plugin-eures` | 8082 |
| Remotive | `plugin-remotive` | 8083 |
| RemoteOK | `plugin-remoteok` | 8084 |

If your service names are different, update the `PLUGIN_*_URL` variables accordingly.

## How to Find Easypanel Service Names

1. Go to Easypanel dashboard
2. Click on each service
3. Look at the service name (usually shown in the URL or header)
4. Use that exact name in the `PLUGIN_*_URL` variables

**Example:**
- If RemoteOK service is named `remoteok-plugin` instead of `plugin-remoteok`
- Set: `PLUGIN_REMOTEOK_URL=http://remoteok-plugin:8084`

## Testing the Setup

### 1. Check Core Can Reach Plugins

From the core container:

```bash
# Test each plugin health endpoint
curl http://plugin-arbetsformedlingen:8081/health
curl http://plugin-eures:8082/health
curl http://plugin-remotive:8083/health
curl http://plugin-remoteok:8084/health
```

Expected response:
```json
{
  "status": "healthy",
  "plugin": "RemoteOK Connector",
  "plugin_id": "remoteok",
  "version": "1.0.0"
}
```

### 2. Trigger Manual Sync

```bash
curl -X POST https://app-openjobs.katsu6.easypanel.host/sync/manual
```

### 3. Check Logs

Look for:
```
🔌 Using HTTP plugin containers (microservices mode)
✅ Arbetsförmedlingen HTTP sync completed
✅ EURES HTTP sync completed
✅ Remotive HTTP sync completed
✅ RemoteOK HTTP sync completed
```

## Troubleshooting

### Plugin URLs Not Working

**Error:** `failed to connect to plugin`

**Solutions:**
1. Check service names match exactly
2. Verify all containers are running
3. Check they're on the same Docker network
4. Try using container IDs instead: `http://<container-id>:8081`

### Still Seeing Duplicates

**Issue:** Connectors running twice

**Solution:** Verify `USE_HTTP_PLUGINS=true` is set in core container

### RemoteOK/Remotive Not Syncing

**Issue:** Only Arbetsförmedlingen and EURES show in logs

**Solution:** 
1. Check plugin containers are running
2. Verify `PLUGIN_REMOTEOK_URL` and `PLUGIN_REMOTIVE_URL` are set
3. Check plugin health endpoints respond
4. Look at plugin container logs for errors

### How to Check Current Environment Variables

In Easypanel:
1. Go to service
2. Click "Environment" tab
3. Verify all variables are set

Or from container:
```bash
docker exec <container-name> env | grep PLUGIN
```

## Deployment Checklist

- [ ] Core container has `USE_HTTP_PLUGINS=true`
- [ ] Core container has all 4 `PLUGIN_*_URL` variables
- [ ] All 5 containers have `SUPABASE_URL` and `SUPABASE_ANON_KEY`
- [ ] All plugin containers are running
- [ ] Service names match the URLs in core container
- [ ] Manual sync test works
- [ ] Dashboard shows all 4 plugins in sync history

## Expected Result

After correct setup, dashboard sync history should show:

```
Plugin                Time          Fetched  Inserted  Duplicates
Arbetsförmedlingen   02:45 18/10   20       0         20
EURES                02:45 18/10   10       0         10
Remotive             02:45 18/10   0        0         0          (SSL issue)
RemoteOK             02:45 18/10   50       30        20
```

Each plugin appears **once per sync cycle**, no duplicates!
