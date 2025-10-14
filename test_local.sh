#!/bin/bash

# OpenJobs Local Testing Script
# This script tests all endpoints and verifies the setup

set -e  # Exit on error

API_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧪 OpenJobs Local Testing Script"
echo "================================"
echo ""

# Check if server is running
echo "1️⃣  Checking if server is running..."
if ! curl -s -f -o /dev/null "$API_URL/health"; then
    echo -e "${RED}❌ Server is not running!${NC}"
    echo "   Please start the server first:"
    echo "   ./openjobs"
    echo ""
    exit 1
fi
echo -e "${GREEN}✅ Server is running${NC}"
echo ""

# Test health endpoint
echo "2️⃣  Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s "$API_URL/health")
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo -e "${GREEN}✅ Health check passed${NC}"
    echo "   Response: $HEALTH_RESPONSE"
else
    echo -e "${RED}❌ Health check failed${NC}"
    echo "   Response: $HEALTH_RESPONSE"
    exit 1
fi
echo ""

# Test plugins endpoint
echo "3️⃣  Testing plugins endpoint..."
PLUGINS_RESPONSE=$(curl -s "$API_URL/plugins")
if echo "$PLUGINS_RESPONSE" | grep -q "success"; then
    echo -e "${GREEN}✅ Plugins endpoint working${NC}"
    echo "   Found $(echo "$PLUGINS_RESPONSE" | grep -o 'arbetsformedlingen-connector' | wc -l) connector(s)"
else
    echo -e "${RED}❌ Plugins endpoint failed${NC}"
    echo "   Response: $PLUGINS_RESPONSE"
    exit 1
fi
echo ""

# Test manual sync
echo "4️⃣  Testing manual job sync..."
echo -e "${YELLOW}   This will fetch jobs from Arbetsförmedlingen API...${NC}"
SYNC_RESPONSE=$(curl -s -X POST "$API_URL/sync/manual")
if echo "$SYNC_RESPONSE" | grep -q "success.*true"; then
    echo -e "${GREEN}✅ Manual sync completed${NC}"
    echo "   Response: $SYNC_RESPONSE"
else
    echo -e "${RED}❌ Manual sync failed${NC}"
    echo "   Response: $SYNC_RESPONSE"
    echo ""
    echo "   Possible issues:"
    echo "   - Supabase credentials not configured correctly"
    echo "   - Database tables not created (run migrations/001_create_job_posts.sql)"
    echo "   - Network connectivity issues"
    exit 1
fi
echo ""

# Wait a moment for jobs to be processed
echo "   Waiting 2 seconds for jobs to be processed..."
sleep 2
echo ""

# Test jobs endpoint
echo "5️⃣  Testing jobs endpoint..."
JOBS_RESPONSE=$(curl -s "$API_URL/jobs?limit=5")
if echo "$JOBS_RESPONSE" | grep -q "success.*true"; then
    JOB_COUNT=$(echo "$JOBS_RESPONSE" | grep -o '"id"' | wc -l)
    if [ "$JOB_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✅ Jobs retrieved successfully${NC}"
        echo "   Found $JOB_COUNT job(s) in the response"
        echo ""
        echo "   Sample job data:"
        echo "$JOBS_RESPONSE" | python3 -m json.tool 2>/dev/null | head -30 || echo "$JOBS_RESPONSE" | head -20
    else
        echo -e "${YELLOW}⚠️  No jobs found in database${NC}"
        echo "   This might be normal on first run or if sync hasn't completed yet"
        echo "   Try running the sync again or wait for automatic sync"
    fi
else
    echo -e "${RED}❌ Jobs endpoint failed${NC}"
    echo "   Response: $JOBS_RESPONSE"
    exit 1
fi
echo ""

# Test creating a manual job
echo "6️⃣  Testing job creation..."
CREATE_RESPONSE=$(curl -s -X POST "$API_URL/jobs" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Test Job from Script",
        "company": "Test Company",
        "description": "This is a test job created by the test script",
        "location": "Stockholm, Sweden",
        "employment_type": "Full-time",
        "experience_level": "Mid-level"
    }')

if echo "$CREATE_RESPONSE" | grep -q "success.*true"; then
    echo -e "${GREEN}✅ Job creation successful${NC}"
    JOB_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "   Created job ID: $JOB_ID"
    
    # Test getting specific job
    if [ -n "$JOB_ID" ]; then
        echo ""
        echo "7️⃣  Testing job retrieval by ID..."
        GET_RESPONSE=$(curl -s "$API_URL/jobs/$JOB_ID")
        if echo "$GET_RESPONSE" | grep -q "Test Job from Script"; then
            echo -e "${GREEN}✅ Job retrieval by ID successful${NC}"
            
            # Clean up test job
            echo ""
            echo "8️⃣  Cleaning up test job..."
            DELETE_RESPONSE=$(curl -s -X DELETE "$API_URL/jobs/$JOB_ID")
            if echo "$DELETE_RESPONSE" | grep -q "success.*true"; then
                echo -e "${GREEN}✅ Test job deleted successfully${NC}"
            else
                echo -e "${YELLOW}⚠️  Could not delete test job${NC}"
            fi
        else
            echo -e "${RED}❌ Job retrieval by ID failed${NC}"
        fi
    fi
else
    echo -e "${RED}❌ Job creation failed${NC}"
    echo "   Response: $CREATE_RESPONSE"
    exit 1
fi
echo ""

# Summary
echo "=========================================="
echo -e "${GREEN}🎉 All tests passed successfully!${NC}"
echo "=========================================="
echo ""
echo "Your OpenJobs installation is working correctly!"
echo ""
echo "Next steps:"
echo "  - Monitor logs for automatic syncs (every 6 hours)"
echo "  - Check your Supabase dashboard to see job data"
echo "  - Deploy to Easypanel using the Dockerfile"
echo ""
echo "Useful commands:"
echo "  - Manual sync:  curl -X POST $API_URL/sync/manual"
echo "  - List jobs:    curl $API_URL/jobs"
echo "  - Health check: curl $API_URL/health"
echo ""