import requests
import random
import json
import time
from datetime import datetime

API_URL = "http://localhost:8081/api/flags"
API_KEY = "test-api-key"
HEADERS = {"X-API-KEY": API_KEY}

def fnv1a_64(data):
    """Python implementation of FNV-1a 64-bit hash (matches Go's hash/fnv)."""
    hash_val = 0xcbf29ce484222325
    prime = 0x100000001b3
    for char in data.encode('utf-8'):
        hash_val ^= char
        hash_val = (hash_val * prime) & 0xffffffffffffffff
    return hash_val

def get_bucket(flag_key, user_id):
    """Calculates deterministic bucket [0-99] for a user."""
    if not user_id: return 100
    h = fnv1a_64(f"{flag_key}:{user_id}")
    return h % 100

def predict_evaluation(flag, context):
    """Local implementation of evaluation logic matching Go backend exactly."""
    if not flag.get("enabled", True):
        return False, "flag disabled"
    
    flag_key = flag.get("key")
    rules = flag.get("rules", [])
    for rule in rules:
        matched, value = eval_rule(rule, flag_key, context)
        if matched:
            return value, f"matched rule: {rule.get('type')}"
            
    return flag.get("defaultValue", False), "default value"

def eval_rule(rule, flag_key, ctx):
    rtype = rule.get("type")
    cfg = rule.get("config", {})
    val = rule.get("value", False)
    user_id = ctx.get("userId", "")

    if rtype == "user_list":
        return (user_id in cfg.get("userIds", [])), val
        
    elif rtype == "percentage":
        if not cfg.get("percentage") or not user_id: return False, False
        bucket = get_bucket(flag_key, user_id)
        return bucket < cfg["percentage"], val

    elif rtype == "gradual":
        if not user_id: return False, False
        try:
            start_at = datetime.fromisoformat(cfg["startAt"].replace("Z", "+00:00"))
            end_at = datetime.fromisoformat(cfg["endAt"].replace("Z", "+00:00"))
            now = datetime.now().astimezone()
            start_p, end_p = cfg.get("startPercent", 0), cfg.get("endPercent", 0)
            if now < start_at: eff_p = start_p
            elif now > end_at: eff_p = end_p
            else:
                progress = (now - start_at).total_seconds() / (end_at - start_at).total_seconds()
                eff_p = start_p + progress * (end_p - start_p)
            bucket = get_bucket(flag_key, user_id)
            return bucket < eff_p, val
        except: return False, False

    elif rtype == "schedule":
        try:
            now = datetime.now().astimezone()
            if cfg.get("enableAt"):
                if now < datetime.fromisoformat(cfg["enableAt"].replace("Z", "+00:00")): return False, False
            if cfg.get("disableAt"):
                if now > datetime.fromisoformat(cfg["disableAt"].replace("Z", "+00:00")): return False, False
            return True, val
        except: return False, False

    elif rtype == "geography":
        # Go logic: If category provided, context MUST match one entry in that category.
        # It's an AND between categories (Country AND State AND City AND Zip).
        if not any([cfg.get("countries"), cfg.get("states"), cfg.get("cities"), cfg.get("zipCodes")]):
            return False, False

        if cfg.get("countries"):
            if not any(c.lower() == ctx.get("country", "").lower() for c in cfg["countries"]):
                return False, False
        if cfg.get("states"):
            if not any(s.lower() == ctx.get("state", "").lower() for s in cfg["states"]):
                return False, False
        if cfg.get("cities"):
            if not any(c.lower() == ctx.get("city", "").lower() for c in cfg["cities"]):
                return False, False
        if cfg.get("zipCodes"):
            if ctx.get("zipCode") not in cfg["zipCodes"]: # Zip is case sensitive in Go? (Wait, checking Go again...)
                return False, False
        return True, val

    elif rtype == "attribute":
        attr_key = cfg.get("attributeKey")
        if not attr_key or attr_key not in ctx.get("attributes", {}): return False, False
        ctx_val = ctx["attributes"][attr_key]
        op, cfg_val = cfg.get("attributeOp"), cfg.get("attributeValue")
        if op == "eq": return (ctx_val == cfg_val), val
        if op == "neq": return (ctx_val != cfg_val), val
        if op == "contains": return (cfg_val in ctx_val), val
        return False, False

    return False, False

def generate_context_for_test(flag, rule, force_match):
    """Generates context to match or mismatch a specific rule."""
    rtype = rule.get("type")
    cfg = rule.get("config", {})
    flag_key = flag.get("key")
    
    ctx = {
        "userId": f"user-{random.randint(1000, 9999)}",
        "country": "US", "state": "CA", "city": "San Francisco", "zipCode": "94105",
        "attributes": {"plan": "free", "tier": "basic"}
    }

    if not force_match:
        ctx["userId"] = "mismatch-user-9999"
        ctx["country"] = "ZZ" # Will fail geography if countries are defined
        ctx["state"] = "MismatchState"
        return ctx

    if rtype == "user_list" and cfg.get("userIds"):
        ctx["userId"] = cfg["userIds"][0]
    elif rtype == "percentage":
        target_p = cfg.get("percentage", 0)
        for i in range(1000):
            uid = f"user-{i}"
            if get_bucket(flag_key, uid) < target_p:
                ctx["userId"] = uid
                break
    elif rtype == "geography":
        # MUST satisfy ALL defined categories
        if cfg.get("countries"): ctx["country"] = cfg["countries"][0]
        if cfg.get("states"): ctx["state"] = cfg["states"][0]
        if cfg.get("cities"): ctx["city"] = cfg["cities"][0]
        if cfg.get("zipCodes"): ctx["zipCode"] = cfg["zipCodes"][0]
    elif rtype == "attribute":
        ctx["attributes"][cfg.get("attributeKey", "plan")] = cfg.get("attributeValue", "pro")
        
    return ctx

def main():
    print(f"[{datetime.now().strftime('%H:%M:%S')}] Fetching flags...")
    try:
        resp = requests.get(API_URL, headers=HEADERS)
        flags = resp.json()
    except Exception as e:
        print(f"Error: {e}"); return

    if not flags:
        print("No flags found."); return

    stats_total = {"passed": 0, "failed": 0}
    stats_types = {}

    print(f"Running 200 tests with FNV-1a bucketing...")

    for i in range(200):
        flag = random.choice(flags)
        rules = flag.get("rules", [])
        target_rule = random.choice(rules) if rules else {"type": "default"}
        rtype = target_rule.get("type")
        
        if rtype not in stats_types: stats_types[rtype] = {"passed": 0, "total": 0}
        stats_types[rtype]["total"] += 1

        context = generate_context_for_test(flag, target_rule, random.choice([True, False]))
        expected_enabled, _ = predict_evaluation(flag, context)
        
        try:
            eval_url = f"{API_URL}/{flag.get('id') or flag.get('_id')}/evaluate"
            resp = requests.post(eval_url, json=context, headers=HEADERS, timeout=2)
            actual_enabled = resp.json().get("enabled") if resp.status_code == 200 else "ERR"
        except:
            actual_enabled = "ERR"

        passed = (actual_enabled == expected_enabled)
        if passed:
            stats_total["passed"] += 1
            stats_types[rtype]["passed"] += 1
        else:
            stats_total["failed"] += 1

    # Summary
    print("\n" + "="*45)
    print(f"{'RULE TYPE':<15} | {'PASS':<6} | {'TOTAL':<6} | {'%'}")
    print("-" * 45)
    for rt, s in sorted(stats_types.items()):
        perc = (s['passed']/s['total'])*100 if s['total'] > 0 else 0
        print(f"{rt:<15} | {s['passed']:<6} | {s['total']:<6} | {perc:.1f}%")
    print("="*45)
    print(f"TOTAL: {stats_total['passed']}/200 ({(stats_total['passed']/200)*100:.1f}%)")

if __name__ == "__main__":
    main()
