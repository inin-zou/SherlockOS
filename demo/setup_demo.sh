#!/bin/bash

# SherlockOS Demo Setup Script
# This script creates a demo case and uploads all the demo data

API_BASE="http://localhost:8080/v1"
DEMO_DIR="$(dirname "$0")"

echo "=== SherlockOS Demo Setup ==="
echo ""

# 1. Create the case
echo "1. Creating demo case..."
CASE_RESPONSE=$(curl -s -X POST "$API_BASE/cases" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Tech Corp Office Break-in",
    "description": "Night-time break-in at Tech Corp headquarters, 3rd floor executive office. Suspect forced entry through window, searched desk and filing cabinets, exited through emergency door."
  }')

# Handle both direct response and wrapped response formats
CASE_ID=$(echo "$CASE_RESPONSE" | jq -r '.data.id // .id')

if [ "$CASE_ID" == "null" ] || [ -z "$CASE_ID" ]; then
  echo "Error creating case: $CASE_RESPONSE"
  exit 1
fi

echo "   Case created: $CASE_ID"
echo ""

# 2. Get upload intents for scene images
echo "2. Getting upload intents for scene images..."
FILES_JSON='{"files":['
FIRST=true
for img in "$DEMO_DIR/images/"*.png; do
  filename=$(basename "$img")
  if [ "$FIRST" = true ]; then
    FIRST=false
  else
    FILES_JSON+=','
  fi
  FILES_JSON+="{\"filename\":\"$filename\",\"content_type\":\"image/png\"}"
done
FILES_JSON+=']}'

UPLOAD_RESPONSE=$(curl -s -X POST "$API_BASE/cases/$CASE_ID/upload-intent" \
  -H "Content-Type: application/json" \
  -d "$FILES_JSON")

echo "   Upload intents received"
BATCH_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.upload_batch_id // .upload_batch_id')
echo "   Batch ID: $BATCH_ID"

# Extract storage keys for later use
STORAGE_KEYS=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.intents[].storage_key // .intents[].storage_key' 2>/dev/null)
echo "   Storage keys:"
echo "$STORAGE_KEYS" | while read key; do
  echo "     - $key"
done
echo ""

# 3. Submit witness statements (triggers profile extraction job)
echo "3. Submitting witness statements..."
STATEMENTS_RESPONSE=$(curl -s -X POST "$API_BASE/cases/$CASE_ID/witness-statements" \
  -H "Content-Type: application/json" \
  -d '{
    "statements": [
      {
        "source_name": "Mike Chen (Security Guard)",
        "content": "I saw a tall man in a dark hoodie run past the east parking lot at approximately 11:45 PM. He was carrying a black backpack and appeared to be limping slightly on his right leg. Height around 6 feet, maybe 6 foot 1. Slim build. I tried to follow but he disappeared around the corner toward Main Street.",
        "credibility": 0.9
      },
      {
        "source_name": "Sarah Johnson (Neighbor)",
        "content": "I was working late in my office on the 2nd floor when I heard glass breaking around 11:30 PM. About 10 minutes later, I looked out my window and saw someone exit through the back emergency door. Medium height, wearing jeans and what looked like a baseball cap. They moved quickly toward the back alley.",
        "credibility": 0.7
      },
      {
        "source_name": "David Park (Delivery Driver)",
        "content": "I was making a late delivery around 11:20 PM. I noticed a white sedan parked oddly near the Tech Corp building. Male driver in his 30s, had a beard, short dark hair. When I walked past, he seemed nervous and looked away. Local license plates.",
        "credibility": 0.6
      },
      {
        "source_name": "Robert Martinez (Night Janitor)",
        "content": "I checked the office at exactly 10 PM. Room 304 was definitely locked and everything looked normal. When I came back at midnight after the alarm, the window was completely shattered and the place was ransacked. I saw muddy footprints leading from the window to the desk and then toward the emergency exit.",
        "credibility": 0.85
      }
    ]
  }')

COMMIT_ID=$(echo "$STATEMENTS_RESPONSE" | jq -r '.data.commit_id // .commit_id')
PROFILE_JOB_ID=$(echo "$STATEMENTS_RESPONSE" | jq -r '.data.profile_job_id // .profile_job_id')
echo "   Statements committed: $COMMIT_ID"
echo "   Profile extraction job: $PROFILE_JOB_ID"
echo ""

# 4. Create a reconstruction job (simulated - using storage keys)
echo "4. Creating scene reconstruction job..."
FIRST_KEY=$(echo "$STORAGE_KEYS" | head -1)
RECON_RESPONSE=$(curl -s -X POST "$API_BASE/cases/$CASE_ID/jobs" \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"reconstruction\",
    \"input\": {
      \"case_id\": \"$CASE_ID\",
      \"scan_asset_keys\": [\"$FIRST_KEY\"]
    }
  }")

RECON_JOB_ID=$(echo "$RECON_RESPONSE" | jq -r '.data.job_id // .job_id')
echo "   Reconstruction job: $RECON_JOB_ID"
echo ""

echo "=== Demo Setup Complete ==="
echo ""
echo "=========================================="
echo "  CASE ID: $CASE_ID"
echo "=========================================="
echo ""
echo "API Endpoints:"
echo "  Case:     $API_BASE/cases/$CASE_ID"
echo "  Timeline: $API_BASE/cases/$CASE_ID/timeline"
echo "  Snapshot: $API_BASE/cases/$CASE_ID/snapshot"
echo ""
echo "Jobs Created:"
echo "  Profile:        $PROFILE_JOB_ID"
echo "  Reconstruction: $RECON_JOB_ID"
echo ""
echo "Next steps:"
echo "  1. Open frontend at http://localhost:3000"
echo "  2. View case timeline and scene"
echo "  3. Check job progress: curl $API_BASE/jobs/$PROFILE_JOB_ID"
echo "  4. Trigger reasoning: POST $API_BASE/cases/$CASE_ID/reasoning"
echo ""
