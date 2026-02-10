#!/bin/bash
# Test script: Upload video to Supabase Storage and trigger a reconstruction job
# Tests the full pipeline: create case -> upload video -> create job -> poll status -> check scene graph
#
# Usage: ./scripts/test_video_reconstruction.sh
#
# Prerequisites:
#   - Backend server running on localhost:8080
#   - Supabase project configured with case-assets bucket
#   - MODAL_MIRROR_URL configured for reconstruction worker

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

API_URL="${API_URL:-http://localhost:8080/v1}"
SUPABASE_URL="${SUPABASE_URL:-https://hdfaugwofzqqdjuzcsin.supabase.co}"
SUPABASE_KEY="${SUPABASE_KEY:-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImhkZmF1Z3dvZnpxcWRqdXpjc2luIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc2OTk1MjQ5NiwiZXhwIjoyMDg1NTI4NDk2fQ.F8vw9cMd_N0M78LBmevrQ6QF-bnYgtoJPgLuE-Kyfa0}"
VIDEO_FILE="${VIDEO_FILE:-/Users/yongkangzou/Desktop/2e7017380d41c87b64f5a80f97ce89b5.mp4}"
POLL_INTERVAL="${POLL_INTERVAL:-5}"
MAX_POLLS="${MAX_POLLS:-120}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info()  { echo -e "${BLUE}[INFO]${NC}  $*"; }
log_ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# =============================================================================
# Step 0: Validate prerequisites
# =============================================================================

echo ""
echo "============================================================"
echo "  SherlockOS - Video Reconstruction Pipeline Test"
echo "============================================================"
echo ""

log_info "API URL:       $API_URL"
log_info "Supabase URL:  $SUPABASE_URL"
log_info "Video file:    $VIDEO_FILE"
echo ""

# Check video file exists
if [ ! -f "$VIDEO_FILE" ]; then
    log_error "Video file not found: $VIDEO_FILE"
    exit 1
fi

VIDEO_SIZE=$(stat -f%z "$VIDEO_FILE" 2>/dev/null || stat -c%s "$VIDEO_FILE" 2>/dev/null || echo "unknown")
VIDEO_SIZE_MB=$(echo "scale=2; $VIDEO_SIZE / 1048576" | bc 2>/dev/null || echo "unknown")
log_info "Video file size: ${VIDEO_SIZE_MB} MB ($VIDEO_SIZE bytes)"

# Check file type
FILE_TYPE=$(file --brief "$VIDEO_FILE")
log_info "Video file type: $FILE_TYPE"

# Check backend is running
if ! curl -sf "$API_URL/../health" > /dev/null 2>&1; then
    # Try without the /v1 prefix
    HEALTH_URL=$(echo "$API_URL" | sed 's|/v1$||')/health
    if ! curl -sf "$HEALTH_URL" > /dev/null 2>&1; then
        log_error "Backend server not reachable. Is it running on $(echo "$API_URL" | sed 's|/v1$||')?"
        log_error "Start it with: cd backend && go run ./cmd/server"
        exit 1
    fi
fi
log_ok "Backend server is reachable"
echo ""

# =============================================================================
# Step 1: Create a test case
# =============================================================================

echo "============================================================"
echo "  Step 1: Create a test case"
echo "============================================================"
echo ""

CASE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/cases" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Video Reconstruction Test",
        "description": "Testing gaussian splatting reconstruction from video input"
    }')

HTTP_CODE=$(echo "$CASE_RESPONSE" | tail -n1)
CASE_BODY=$(echo "$CASE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" != "201" ]; then
    log_error "Failed to create case (HTTP $HTTP_CODE)"
    echo "$CASE_BODY" | python3 -m json.tool 2>/dev/null || echo "$CASE_BODY"
    exit 1
fi

CASE_ID=$(echo "$CASE_BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
log_ok "Case created: $CASE_ID"
echo "$CASE_BODY" | python3 -m json.tool 2>/dev/null || echo "$CASE_BODY"
echo ""

# =============================================================================
# Step 2: Upload video to Supabase Storage
# =============================================================================

echo "============================================================"
echo "  Step 2: Upload video to Supabase Storage"
echo "============================================================"
echo ""

STORAGE_KEY="cases/${CASE_ID}/scans/input_video.mp4"
UPLOAD_URL="${SUPABASE_URL}/storage/v1/object/case-assets/${STORAGE_KEY}"

log_info "Uploading to: $STORAGE_KEY"
log_info "Upload URL:   $UPLOAD_URL"
log_info "This may take a while for large files..."
echo ""

UPLOAD_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$UPLOAD_URL" \
    -H "apikey: $SUPABASE_KEY" \
    -H "Authorization: Bearer $SUPABASE_KEY" \
    -H "Content-Type: video/mp4" \
    -H "x-upsert: true" \
    --data-binary @"$VIDEO_FILE")

UPLOAD_HTTP_CODE=$(echo "$UPLOAD_RESPONSE" | tail -n1)
UPLOAD_BODY=$(echo "$UPLOAD_RESPONSE" | sed '$d')

if [ "$UPLOAD_HTTP_CODE" = "200" ] || [ "$UPLOAD_HTTP_CODE" = "201" ]; then
    log_ok "Video uploaded successfully (HTTP $UPLOAD_HTTP_CODE)"
else
    log_error "Upload failed (HTTP $UPLOAD_HTTP_CODE)"
    echo "$UPLOAD_BODY" | python3 -m json.tool 2>/dev/null || echo "$UPLOAD_BODY"
    exit 1
fi

echo "$UPLOAD_BODY" | python3 -m json.tool 2>/dev/null || echo "$UPLOAD_BODY"
echo ""

# Verify the upload by checking the object exists
VERIFY_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $SUPABASE_KEY" \
    "${SUPABASE_URL}/storage/v1/object/info/case-assets/${STORAGE_KEY}" 2>/dev/null || echo "000")

if [ "$VERIFY_RESPONSE" = "200" ]; then
    log_ok "Upload verified - file exists in storage"
else
    log_warn "Could not verify upload (HTTP $VERIFY_RESPONSE) - continuing anyway"
fi
echo ""

# =============================================================================
# Step 3: Trigger reconstruction job
# =============================================================================

echo "============================================================"
echo "  Step 3: Trigger reconstruction job"
echo "============================================================"
echo ""

log_info "Creating reconstruction job with video asset key: $STORAGE_KEY"
echo ""

JOB_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/cases/$CASE_ID/jobs" \
    -H "Content-Type: application/json" \
    -H "Idempotency-Key: test-video-recon-$(date +%s)" \
    -d "{
        \"type\": \"reconstruction\",
        \"input\": {
            \"case_id\": \"$CASE_ID\",
            \"video_asset_key\": \"$STORAGE_KEY\"
        }
    }")

JOB_HTTP_CODE=$(echo "$JOB_RESPONSE" | tail -n1)
JOB_BODY=$(echo "$JOB_RESPONSE" | sed '$d')

if [ "$JOB_HTTP_CODE" = "202" ] || [ "$JOB_HTTP_CODE" = "200" ]; then
    log_ok "Reconstruction job created (HTTP $JOB_HTTP_CODE)"
elif [ "$JOB_HTTP_CODE" = "503" ]; then
    log_error "Reconstruction service unavailable (HTTP 503)"
    log_error "Make sure MODAL_MIRROR_URL is configured and the reconstruction worker is registered."
    echo "$JOB_BODY" | python3 -m json.tool 2>/dev/null || echo "$JOB_BODY"
    exit 1
else
    log_error "Failed to create job (HTTP $JOB_HTTP_CODE)"
    echo "$JOB_BODY" | python3 -m json.tool 2>/dev/null || echo "$JOB_BODY"
    exit 1
fi

JOB_ID=$(echo "$JOB_BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['job_id'])")
log_ok "Job ID: $JOB_ID"
echo "$JOB_BODY" | python3 -m json.tool 2>/dev/null || echo "$JOB_BODY"
echo ""

# =============================================================================
# Step 4: Poll job status
# =============================================================================

echo "============================================================"
echo "  Step 4: Poll job status (every ${POLL_INTERVAL}s, max ${MAX_POLLS} polls)"
echo "============================================================"
echo ""

# Note: job GET endpoint is /v1/jobs/{jobId} (not nested under cases)
FINAL_STATUS="unknown"
for i in $(seq 1 "$MAX_POLLS"); do
    STATUS_RESPONSE=$(curl -s "$API_URL/jobs/$JOB_ID")
    JOB_STATUS=$(echo "$STATUS_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['status'])" 2>/dev/null || echo "unknown")
    PROGRESS=$(echo "$STATUS_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['progress'])" 2>/dev/null || echo "0")

    TIMESTAMP=$(date +"%H:%M:%S")
    echo -e "  [$TIMESTAMP] Poll $i/$MAX_POLLS | Status: ${YELLOW}${JOB_STATUS}${NC} | Progress: ${PROGRESS}%"

    if [ "$JOB_STATUS" = "done" ]; then
        echo ""
        log_ok "Job completed successfully!"
        FINAL_STATUS="done"

        # Print job output
        echo ""
        log_info "Job output:"
        echo "$STATUS_RESPONSE" | python3 -c "
import sys, json
data = json.load(sys.stdin)['data']
output = data.get('output', {})
if isinstance(output, str):
    import json as j
    output = j.loads(output)

# Print key output fields
gk = output.get('gaussian_asset_key', '')
mk = output.get('mesh_asset_key', '')
pk = output.get('pointcloud_asset_key', '')
stats = output.get('processing_stats', {})
objects = output.get('objects', [])

print(f'  Gaussian Asset Key: {gk}')
print(f'  Mesh Asset Key:     {mk}')
print(f'  Pointcloud Key:     {pk}')
print(f'  Objects Detected:   {len(objects)}')
print(f'  Processing Stats:   {json.dumps(stats, indent=4)}')
" 2>/dev/null || echo "$STATUS_RESPONSE" | python3 -m json.tool 2>/dev/null
        break
    fi

    if [ "$JOB_STATUS" = "failed" ]; then
        echo ""
        log_error "Job failed!"
        FINAL_STATUS="failed"
        JOB_ERROR=$(echo "$STATUS_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'].get('error','unknown'))" 2>/dev/null || echo "unknown")
        log_error "Error: $JOB_ERROR"
        echo ""
        echo "Full response:"
        echo "$STATUS_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$STATUS_RESPONSE"
        break
    fi

    if [ "$JOB_STATUS" = "canceled" ]; then
        echo ""
        log_warn "Job was canceled"
        FINAL_STATUS="canceled"
        break
    fi

    sleep "$POLL_INTERVAL"
done

if [ "$FINAL_STATUS" = "unknown" ]; then
    log_warn "Polling timed out after $((MAX_POLLS * POLL_INTERVAL)) seconds"
    log_warn "Job may still be running. Check manually:"
    log_warn "  curl -s $API_URL/jobs/$JOB_ID | python3 -m json.tool"
fi
echo ""

# =============================================================================
# Step 5: Check scene graph for gaussian_asset_key
# =============================================================================

echo "============================================================"
echo "  Step 5: Check scene graph (snapshot)"
echo "============================================================"
echo ""

SNAPSHOT_RESPONSE=$(curl -s "$API_URL/cases/$CASE_ID/snapshot")
echo "$SNAPSHOT_RESPONSE" | python3 -c "
import sys, json

data = json.load(sys.stdin)
if not data.get('success'):
    print('Failed to get snapshot:', data.get('error', {}).get('message', 'unknown'))
    sys.exit(0)

snapshot = data.get('data', {})
sg = snapshot.get('scenegraph', {})

print('Scene Graph Summary:')
print(f'  Version:        {sg.get(\"version\", \"N/A\")}')
print(f'  Objects:        {len(sg.get(\"objects\", []))}')
print(f'  Evidence cards: {len(sg.get(\"evidence\", []))}')
print(f'  Constraints:    {len(sg.get(\"constraints\", []))}')
print(f'  Uncertainty:    {len(sg.get(\"uncertainty_regions\", []))}')
print()

gk = sg.get('gaussian_asset_key', '')
if gk:
    print(f'  GAUSSIAN ASSET KEY: {gk}')
    print(f'  Public URL: https://hdfaugwofzqqdjuzcsin.supabase.co/storage/v1/object/public/case-assets/{gk}')
else:
    print('  No gaussian_asset_key found in scene graph')

pc = sg.get('point_cloud')
if pc:
    print(f'  Point cloud: {pc.get(\"count\", 0)} points')

# List objects
objects = sg.get('objects', [])
if objects:
    print()
    print('  Detected Objects:')
    for obj in objects:
        print(f'    - [{obj.get(\"type\",\"?\")}] {obj.get(\"label\",\"unnamed\")} (confidence: {obj.get(\"confidence\",0):.2f})')
" 2>/dev/null || echo "$SNAPSHOT_RESPONSE" | python3 -m json.tool 2>/dev/null
echo ""

# =============================================================================
# Step 6: Summary
# =============================================================================

echo "============================================================"
echo "  Summary"
echo "============================================================"
echo ""

log_info "Case ID:     $CASE_ID"
log_info "Job ID:      $JOB_ID"
log_info "Job Status:  $FINAL_STATUS"
log_info "Video Key:   $STORAGE_KEY"
echo ""

if [ "$FINAL_STATUS" = "done" ]; then
    log_ok "Pipeline test PASSED"
else
    log_warn "Pipeline test ended with status: $FINAL_STATUS"
fi

echo ""
log_info "Frontend URL:  http://localhost:3000/cases/$CASE_ID"
log_info "Job API:       $API_URL/jobs/$JOB_ID"
log_info "Snapshot API:  $API_URL/cases/$CASE_ID/snapshot"
echo ""
echo "============================================================"
echo "  Done"
echo "============================================================"
