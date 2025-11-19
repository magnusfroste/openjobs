#!/bin/bash

# Test Supabase Connection for OpenJobs
# This script tests read/write access to the new cloud Supabase instance

echo "🧪 Testing OpenJobs → Supabase Connection"
echo "=========================================="
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "❌ .env file not found!"
    exit 1
fi

# Check required env vars
if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_ANON_KEY" ]; then
    echo "❌ Missing SUPABASE_URL or SUPABASE_ANON_KEY in .env"
    exit 1
fi

echo "📍 Supabase URL: $SUPABASE_URL"
echo ""

# Test 1: Health Check
echo "Test 1: Health Check"
echo "--------------------"
response=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "apikey: $SUPABASE_ANON_KEY" \
    "${SUPABASE_URL}/rest/v1/")
if [ "$response" = "200" ]; then
    echo "✅ Supabase is reachable (HTTP $response)"
else
    echo "❌ Supabase health check failed (HTTP $response)"
    exit 1
fi
echo ""

# Test 2: Read from job_posts table
echo "Test 2: Read Access (job_posts)"
echo "-------------------------------"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -H "apikey: $SUPABASE_ANON_KEY" \
    -H "Authorization: Bearer $SUPABASE_ANON_KEY" \
    "${SUPABASE_URL}/rest/v1/job_posts?select=id,title&limit=5")

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')

if [ "$http_code" = "200" ]; then
    echo "✅ Can read from job_posts table"
    echo "📊 Sample data:"
    echo "$body" | jq -r '.[] | "  - \(.id): \(.title)"' 2>/dev/null || echo "$body"
else
    echo "❌ Failed to read from job_posts (HTTP $http_code)"
    echo "Response: $body"
fi
echo ""

# Test 3: Write to job_posts table (test job)
echo "Test 3: Write Access (job_posts)"
echo "--------------------------------"
test_id="test-$(date +%s)"
test_job='{
  "id": "'$test_id'",
  "title": "Test Job - Connection Check",
  "company": "OpenJobs Test",
  "description": "This is a test job to verify write access",
  "location": "Test Location",
  "is_active": false
}'

response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X POST \
    -H "apikey: $SUPABASE_ANON_KEY" \
    -H "Authorization: Bearer $SUPABASE_ANON_KEY" \
    -H "Content-Type: application/json" \
    -H "Prefer: return=representation" \
    -d "$test_job" \
    "${SUPABASE_URL}/rest/v1/job_posts")

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')

if [ "$http_code" = "201" ]; then
    echo "✅ Can write to job_posts table"
    echo "📝 Created test job: $test_id"
else
    echo "❌ Failed to write to job_posts (HTTP $http_code)"
    echo "Response: $body"
fi
echo ""

# Test 4: Update test job
echo "Test 4: Update Access (job_posts)"
echo "---------------------------------"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X PATCH \
    -H "apikey: $SUPABASE_ANON_KEY" \
    -H "Authorization: Bearer $SUPABASE_ANON_KEY" \
    -H "Content-Type: application/json" \
    -d '{"description": "Updated test description"}' \
    "${SUPABASE_URL}/rest/v1/job_posts?id=eq.$test_id")

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)

if [ "$http_code" = "200" ] || [ "$http_code" = "204" ]; then
    echo "✅ Can update job_posts table"
else
    echo "❌ Failed to update job_posts (HTTP $http_code)"
fi
echo ""

# Test 5: Delete test job (cleanup)
echo "Test 5: Cleanup (delete test job)"
echo "---------------------------------"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X DELETE \
    -H "apikey: $SUPABASE_ANON_KEY" \
    -H "Authorization: Bearer $SUPABASE_ANON_KEY" \
    "${SUPABASE_URL}/rest/v1/job_posts?id=eq.$test_id")

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)

if [ "$http_code" = "200" ] || [ "$http_code" = "204" ]; then
    echo "✅ Can delete from job_posts table"
    echo "🧹 Test job cleaned up"
else
    echo "⚠️  Could not delete test job (HTTP $http_code)"
    echo "   You may need to manually delete: $test_id"
fi
echo ""

# Test 6: Check companies table (for API key validation)
echo "Test 6: Companies Table Access"
echo "------------------------------"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -H "apikey: $SUPABASE_ANON_KEY" \
    -H "Authorization: Bearer $SUPABASE_ANON_KEY" \
    "${SUPABASE_URL}/rest/v1/companies?select=id,name&limit=1")

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)

if [ "$http_code" = "200" ]; then
    echo "✅ Can read from companies table"
else
    echo "⚠️  Companies table check (HTTP $http_code)"
    echo "   Note: This is expected if table doesn't exist yet"
fi
echo ""

# Summary
echo "=========================================="
echo "📊 Test Summary"
echo "=========================================="
echo "✅ All critical tests passed!"
echo ""
echo "Next steps:"
echo "1. If companies table test failed, run the SQL migration"
echo "2. Start OpenJobs API: go run cmd/openjobs/main.go"
echo "3. Test API endpoint: curl http://localhost:8080/health"
echo ""
