import requests
import json

API_URL = "http://localhost:8081/api/flags"
API_KEY = "test-api-key"
headers = {"X-API-KEY": API_KEY}

# Step 1: List all flags
response = requests.get(API_URL, headers=headers)
if response.status_code != 200:
    print(f"Failed to fetch flags: {response.status_code} - {response.text}")
    exit(1)

flags = response.json()
print(f"Found {len(flags)} flags.")

# Step 2: For each flag, call get-by-id endpoint to see full details
for flag in flags:
    flag_id = flag.get("id") or flag.get("_id")
    if not flag_id:
        print(f"Flag missing id: {flag}")
        continue
    
    get_url = f"{API_URL}/{flag_id}"
    resp = requests.get(get_url, headers=headers)
    if resp.status_code == 200:
        f = resp.json()
        strategy = f.get("ruleMatchStrategy", "any")
        num_rules = len(f.get("rules", []))
        print(f"ID: {flag_id} | Key: {f['key']} | Enabled: {f['enabled']} | Rules: {num_rules} | Strategy: {strategy}")
    else:
        print(f"Failed to fetch flag {flag_id}: {resp.status_code} - {resp.text}")
