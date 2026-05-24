import requests
import json
import sys

# Configuration
BASE_URL = "http://localhost:8081"
API_KEY = "test-api-key"
FLAG_KEY = "OnOffFeatureFlag"

def create_flag():
    url = f"{BASE_URL}/api/flags"
    headers = {
        "X-API-KEY": API_KEY,
        "Content-Type": "application/json"
    }
    payload = {
        "key": FLAG_KEY,
        "name": "On Off Feature Flag",
        "description": "A simple on/off flag with no rules",
        "enabled": True,
        "defaultValue": True,
        "rules": [],
        "createdBy": "script-user"
    }
    
    print(f"Creating flag: {FLAG_KEY}...")
    response = requests.post(url, headers=headers, json=payload)
    
    if response.status_code == 201:
        print("Flag created successfully.")
        return response.json()
    elif response.status_code == 409:
        print(f"Flag with key '{FLAG_KEY}' already exists. Skipping creation.")
        return None
    else:
        print(f"Failed to create flag. Status: {response.status_code}, Body: {response.text}")
        response.raise_for_status()

def evaluate_flag():
    url = f"{BASE_URL}/api/flags/{FLAG_KEY}/evaluate"
    headers = {
        "X-API-KEY": API_KEY,
        "Content-Type": "application/json"
    }
    # No attributes needed for this simple flag
    payload = {
        "userId": "test-user"
    }
    
    print(f"Evaluating flag: {FLAG_KEY}...")
    response = requests.post(url, headers=headers, json=payload)
    
    if response.status_code == 200:
        result = response.json()
        enabled = result.get("enabled")
        print(f"Evaluation result: {enabled}")
        if enabled is True:
            print("Validation PASSED: Flag is enabled.")
        else:
            print("Validation FAILED: Flag is disabled.")
    else:
        print(f"Failed to evaluate flag. Status: {response.status_code}, Body: {response.text}")
        response.raise_for_status()

if __name__ == "__main__":
    try:
        create_flag()
        evaluate_flag()
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
