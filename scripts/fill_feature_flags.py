import random
import string
import time
from datetime import datetime, timedelta
import requests

def random_flag_name():
    return 'flag_' + ''.join(random.choices(string.ascii_lowercase, k=8))

def random_flag_key():
    return 'key_' + ''.join(random.choices(string.ascii_lowercase + string.digits, k=8))

def random_rule():
    rule_types = [
        'percentage', 'gradual', 'geography', 'schedule', 'user_list', 'attribute'
    ]
    rule_type = random.choice(rule_types)
    config = {}
    now = datetime.utcnow()
    if rule_type == 'percentage':
        config['percentage'] = round(random.uniform(1, 100), 2)
    elif rule_type == 'gradual':
        start = now
        end = now + timedelta(days=random.randint(1, 30))
        config['startPercent'] = round(random.uniform(0, 50), 2)
        config['endPercent'] = round(random.uniform(50, 100), 2)
        config['startAt'] = start.isoformat() + 'Z'
        config['endAt'] = end.isoformat() + 'Z'
    elif rule_type == 'geography':
        config['countries'] = [random.choice(['US', 'CA', 'GB', 'DE', 'FR', 'IN', 'JP'])]
        config['states'] = [random.choice(['CA', 'NY', 'TX', 'ON', 'BC', 'Bavaria', 'Île-de-France'])]
        config['cities'] = [random.choice(['San Francisco', 'Toronto', 'London', 'Berlin', 'Paris', 'Tokyo'])]
        config['zipCodes'] = [str(random.randint(10000, 99999))]
    elif rule_type == 'schedule':
        enable_at = now + timedelta(minutes=random.randint(1, 60))
        disable_at = enable_at + timedelta(hours=random.randint(1, 48))
        config['enableAt'] = enable_at.isoformat() + 'Z'
        config['disableAt'] = disable_at.isoformat() + 'Z'
    elif rule_type == 'user_list':
        config['userIds'] = [f'user-{random.randint(1, 100)}' for _ in range(random.randint(1, 3))]
    elif rule_type == 'attribute':
        key = random.choice(['plan', 'age', 'country', 'tier'])
        op = random.choice(['eq', 'neq', 'contains', 'gt', 'lt'])
        if key == 'age':
            value = str(random.randint(18, 65))
        elif key == 'plan':
            value = random.choice(['free', 'pro', 'enterprise'])
        elif key == 'country':
            value = random.choice(['US', 'CA', 'GB', 'DE', 'FR', 'IN', 'JP'])
        else:
            value = random.choice(['basic', 'premium'])
        config['attributeKey'] = key
        config['attributeOp'] = op
        config['attributeValue'] = value
    return {
        'description': f'Random {rule_type} rule',
        'type': rule_type,
        'config': config,
        'value': random.choice([True, False])
    }

def random_flag():
    now = datetime.utcnow()
    return {
        'name': random_flag_name(),
        'key': random_flag_key(),
        'enabled': random.choice([True, False]),
        'description': 'Randomly generated feature flag',
        'offValue': random.choice([True, False]),
        'fallthroughValue': random.choice([True, False]),
        'rules': [random_rule() for _ in range(random.randint(1, 3))],
        'ruleMatchStrategy': random.choice(['any', 'all']),
        'createdAt': now,
        'createdBy': 'seed-script',
        'updatedAt': now,
        'updatedBy': 'seed-script',
    }


def insert_random_flags(n=20, api_url="http://localhost:8081/api/flags"):
    flags = [random_flag() for _ in range(n)]
    success = 0
    headers = {"X-API-KEY": "test-api-key"}
    for flag in flags:
        # The API expects the CreateFlagRequest structure
        payload = {
            "key": flag["key"],
            "name": flag["name"],
            "enabled": flag["enabled"],
            "description": flag["description"],
            "offValue": flag["offValue"],
            "fallthroughValue": flag["fallthroughValue"],
            "rules": flag["rules"],
            "ruleMatchStrategy": flag["ruleMatchStrategy"],
            "createdBy": flag["createdBy"]
        }
        resp = requests.post(api_url, json=payload, headers=headers)
        if resp.status_code == 201:
            print(f"Inserted flag: {flag['key']} (strategy: {flag['ruleMatchStrategy']})")
            success += 1
        else:
            print(f"Failed to insert flag: {flag['key']} (status {resp.status_code}) - {resp.text}")
    print(f"Inserted {success}/{n} feature flags via API.")

if __name__ == '__main__':
    insert_random_flags(20)
