#!/bin/bash

# Apply migration 007 to fix zero timestamps
# This script uses Supabase REST API to update jobs with zero timestamps

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

SUPABASE_URL="${SUPABASE_URL:-https://cmpnqpdxhmecptcbffmw.supabase.co}"
SUPABASE_KEY="${SUPABASE_ANON_KEY}"

if [ -z "$SUPABASE_KEY" ]; then
    echo "❌ Error: SUPABASE_ANON_KEY not set"
    exit 1
fi

echo "🔧 Applying migration 007: Fix zero timestamps"
echo "📍 Supabase URL: $SUPABASE_URL"
echo ""

# Step 1: Check how many jobs have zero timestamps
echo "📊 Checking jobs with zero timestamps..."
ZERO_COUNT=$(curl -s "$SUPABASE_URL/rest/v1/job_posts?select=count&created_at=lt.2000-01-01" \
    -H "apikey: $SUPABASE_KEY" \
    -H "Prefer: count=exact" | jq -r '.[0].count')

echo "   Found $ZERO_COUNT jobs with zero timestamps"
echo ""

if [ "$ZERO_COUNT" -eq 0 ]; then
    echo "✅ No jobs to fix!"
    exit 0
fi

# Step 2: Get the jobs that need fixing
echo "📋 Fetching jobs to fix..."
JOBS=$(curl -s "$SUPABASE_URL/rest/v1/job_posts?select=id,posted_date&created_at=lt.2000-01-01" \
    -H "apikey: $SUPABASE_KEY")

echo "   Retrieved $(echo $JOBS | jq '. | length') jobs"
echo ""

# Step 3: Update each job (since we can't do complex SQL via REST API)
echo "🔄 Updating jobs..."
UPDATED=0
FAILED=0

echo "$JOBS" | jq -c '.[]' | while read -r job; do
    ID=$(echo $job | jq -r '.id')
    POSTED_DATE=$(echo $job | jq -r '.posted_date')
    
    # Use posted_date as created_at and updated_at
    RESULT=$(curl -s -X PATCH "$SUPABASE_URL/rest/v1/job_posts?id=eq.$ID" \
        -H "apikey: $SUPABASE_KEY" \
        -H "Content-Type: application/json" \
        -H "Prefer: return=minimal" \
        -d "{\"created_at\": \"$POSTED_DATE\", \"updated_at\": \"$POSTED_DATE\"}")
    
    if [ $? -eq 0 ]; then
        echo "   ✅ Updated job: $ID"
        ((UPDATED++))
    else
        echo "   ❌ Failed to update job: $ID"
        ((FAILED++))
    fi
done

echo ""
echo "📊 Summary:"
echo "   Total jobs with zero timestamps: $ZERO_COUNT"
echo "   Successfully updated: $UPDATED"
echo "   Failed: $FAILED"
echo ""

# Step 4: Verify the fix
echo "🔍 Verifying fix..."
REMAINING=$(curl -s "$SUPABASE_URL/rest/v1/job_posts?select=count&created_at=lt.2000-01-01" \
    -H "apikey: $SUPABASE_KEY" \
    -H "Prefer: count=exact" | jq -r '.[0].count')

echo "   Jobs with zero timestamps remaining: $REMAINING"
echo ""

if [ "$REMAINING" -eq 0 ]; then
    echo "✅ Migration 007 applied successfully!"
else
    echo "⚠️  Warning: $REMAINING jobs still have zero timestamps"
fi
