#!/bin/bash

# --- CONFIG ---
API_URL="http://localhost:8081/api/flags"
API_KEY="test-api-key"
CONCURRENCY=2000
DURATION="60s"
# --------------

echo "--- BOMBARDIER STRESS TEST (KEY-BASED) ---"

# 1. Check for bombardier
if ! command -v bombardier &> /dev/null; then
    echo "bombardier not found. Installing..."
    go install github.com/codesenberg/bombardier@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# 2. Get a random enabled flag KEY
echo "Fetching flags..."
FLAGS=$(curl -s -H "X-API-KEY: $API_KEY" "$API_URL")
FLAG_KEY=$(echo "$FLAGS" | jq -r '[.[] | select(.enabled == true)] | .[0].key')

if [ "$FLAG_KEY" == "null" ] || [ -z "$FLAG_KEY" ]; then
    echo "No enabled flags found. Please populate DB first."
    exit 1
fi

echo "Targeting Flag KEY: $FLAG_KEY"
echo "Concurrency: $CONCURRENCY | Duration: $DURATION"
echo "-----------------------------------"

# 3. Slam it
bombardier -c "$CONCURRENCY" -d "$DURATION" -m POST \
    -H "X-API-KEY: $API_KEY" \
    -H "Content-Type: application/json" \
    -b '{"userId":"stress-test-user","attributes":{"tier":"gold","source":"bombardier"}}' \
    "$API_URL/$FLAG_KEY/evaluate"
