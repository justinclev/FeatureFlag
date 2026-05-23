import requests
import random

API_URL = "http://localhost:8081/api/flags"
API_KEY = "test-api-key"
HEADERS = {"X-API-KEY": API_KEY}

def generate_random_context():
    """Generates a random evaluation context matching the API schema."""
    return {
        "userId": f"user-{random.randint(1, 100)}",
        "country": random.choice(["US", "CA", "GB", "DE", "FR", "IN", "JP"]),
        "state": random.choice(["CA", "NY", "TX", "ON", "BC", "Bavaria", "Île-de-France"]),
        "city": random.choice(["San Francisco", "Toronto", "London", "Berlin", "Paris", "Tokyo"]),
        "zipCode": str(random.randint(10000, 99999)),
        "attributes": {
            "plan": random.choice(["free", "pro", "enterprise"]),
            "age": str(random.randint(18, 65)),
            "tier": random.choice(["basic", "premium"])
        }
    }

def main():
    # Step 1: List all flags
    print("Fetching list of flags...")
    response = requests.get(API_URL, headers=HEADERS)
    if response.status_code != 200:
        print(f"Failed to fetch flags: {response.status_code} - {response.text}")
        exit(1)

    flags = response.json()
    if not flags:
        print("No flags found. Try running fill_feature_flags.py first.")
        exit(0)

    print(f"Found {len(flags)} flags. Evaluating them with random contexts...\n")

    # Step 2: Evaluate flags
    success_count = 0
    for flag in flags:
        flag_id = flag.get("id") or flag.get("_id")
        flag_key = flag.get("key", "unknown")
        
        if not flag_id:
            print(f"Flag missing id: {flag_key}")
            continue

        eval_url = f"{API_URL}/{flag_id}/evaluate"
        context = generate_random_context()

        resp = requests.post(eval_url, json=context, headers=HEADERS)
        if resp.status_code == 200:
            result = resp.json()
            enabled = result.get("enabled")
            reason = result.get("reason")
            print(f"[OK] Flag '{flag_key}': enabled={enabled} | reason='{reason}'")
            success_count += 1
        else:
            print(f"[ERROR] Failed to evaluate flag '{flag_key}': {resp.status_code} - {resp.text}")

    print(f"\nSuccessfully evaluated {success_count} out of {len(flags)} flags.")

if __name__ == '__main__':
    main()
