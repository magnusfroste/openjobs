# Fix: Invalid Cron Schedule Error

## Error Message
```
❌ Invalid cron schedule '00 4 * * *.': failed to parse int from *.: strconv.Atoi: parsing "*.": invalid syntax
```

## Problem
The `CRON_SCHEDULE` environment variable in Easypanel has an extra period (`.`) at the end:
- ❌ Wrong: `00 4 * * *.`
- ✅ Correct: `00 4 * * *`

## Solution

### Fix in Easypanel

1. **Go to Easypanel Dashboard**
   - Navigate to your OpenJobs app
   - Click on "Environment" tab

2. **Find CRON_SCHEDULE Variable**
   - Look for: `CRON_SCHEDULE=00 4 * * *.`

3. **Remove the Period**
   - Change to: `CRON_SCHEDULE=00 4 * * *`
   - Or for 6 AM: `CRON_SCHEDULE=0 6 * * *`

4. **Restart the Container**
   - Click "Restart" button
   - Wait for container to come back up

### Verify Fix

Check logs for:
```
✅ ⏰ Starting job ingestion with cron schedule: 00 4 * * *
✅ Cron scheduler started
```

Instead of:
```
❌ Invalid cron schedule '00 4 * * *.': failed to parse int from *.: strconv.Atoi: parsing "*.": invalid syntax
```

## Cron Schedule Examples

```bash
# Every day at 6:00 AM
CRON_SCHEDULE=0 6 * * *

# Every day at 4:00 AM (UTC)
CRON_SCHEDULE=0 4 * * *

# Every 6 hours
CRON_SCHEDULE=0 */6 * * *

# Every day at midnight
CRON_SCHEDULE=0 0 * * *
```

## Format
```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
│ │ │ │ │
* * * * *
```

## Note
This error does NOT affect local development since `CRON_SCHEDULE` is not set in `.env` file.
It only affects the Easypanel deployment.
