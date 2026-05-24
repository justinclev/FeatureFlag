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

# Step 2: Display summary info
for f in flags:
    strategy = f.get("ruleMatchStrategy", "any")
    num_rules = len(f.get("rules", []))
    print(f"Key: {f['key']:<20} | Enabled: {str(f['enabled']):<5} | Rules: {num_rules} | Strategy: {strategy} | ID: {f['id']}")
