import requests
import random
import json
import time
import math
from datetime import datetime, timezone

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
    """Calculates deterministic bucket [0-99.99] for a user with 0.01 precision."""
    if not user_id: return 100.0
    h = fnv1a_64(f"{flag_key}:{user_id}")
    return (h % 10000) / 100.0

def to_string(v):
    """Matches Go backend's toString(v any) logic exactly."""
    if v is None: return ""
    if isinstance(v, bool): return str(v).lower()
    if isinstance(v, (int, float)):
        s = f"{v:g}"
        return s
    return str(v)

def predict_evaluation(flag, context, server_now=None):
    """Local implementation of evaluation logic matching Go backend exactly."""
    if not flag.get("enabled", True):
        return False, "flag disabled"
    
    rules = flag.get("rules", [])
    if not rules:
        return flag.get("defaultValue", False), "default value (no rules)"

    strategy = flag.get("ruleMatchStrategy", "any")

    if strategy == "all":
        last_value = False
        for rule in rules:
            matched, value = eval_rule(rule, flag.get("key"), context, server_now)
            if not matched:
                return flag.get("defaultValue", False), f"failed rule: {rule.get('type')}"
            last_value = value
        return last_value, "matched all rules"
    else:
        # Default: ANY
        for rule in rules:
            matched, value = eval_rule(rule, flag.get("key"), context, server_now)
            if matched:
                return value, f"matched rule: {rule.get('type')}"
        return flag.get("defaultValue", False), "default value"

def eval_rule(rule, flag_key, ctx, server_now=None):
    rtype = rule.get("type")
    cfg = rule.get("config", {})
    val = rule.get("value", False)
    user_id = ctx.get("userId", "")

    if rtype == "user_list":
        # Synchronized logic: Convert all user_ids to strings for comparison
        uids = [to_string(uid).lower().strip() for uid in cfg.get("userIds", [])]
        return (to_string(user_id).lower().strip() in uids), val
        
    elif rtype == "percentage":
        p = cfg.get("percentage")
        if p is None or not user_id: return False, False
        bucket = get_bucket(flag_key, user_id)
        return bucket < float(p), val

    elif rtype == "gradual":
        if not user_id: return False, False
        try:
            start_at = datetime.fromisoformat(cfg["startAt"].replace("Z", "+00:00")).astimezone(timezone.utc)
            end_at = datetime.fromisoformat(cfg["endAt"].replace("Z", "+00:00")).astimezone(timezone.utc)
            now = server_now if server_now else datetime.now(timezone.utc)
            
            start_p = float(cfg.get("startPercent", 0))
            end_p = float(cfg.get("endPercent", 0))
            
            if now < start_at: eff_p = start_p
            elif now > end_at: eff_p = end_p
            else:
                duration = (end_at - start_at).total_seconds()
                if duration <= 0:
                    eff_p = end_p
                else:
                    progress = (now - start_at).total_seconds() / duration
                    eff_p = start_p + progress * (end_p - start_p)
            bucket = get_bucket(flag_key, user_id)
            return bucket < eff_p, val
        except: return False, False

    elif rtype == "schedule":
        try:
            now = server_now if server_now else datetime.now(timezone.utc)
            if cfg.get("enableAt"):
                ea = datetime.fromisoformat(cfg["enableAt"].replace("Z", "+00:00")).astimezone(timezone.utc)
                if now < ea: return False, False
            if cfg.get("disableAt"):
                da = datetime.fromisoformat(cfg["disableAt"].replace("Z", "+00:00")).astimezone(timezone.utc)
                if now > da: return False, False
            return True, val
        except: return False, False

    elif rtype == "geography":
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
            if any(z.lower() == ctx.get("zipCode", "").lower() for z in cfg["zipCodes"]):
                return True, val
            return False, False
        return True, val

    elif rtype == "attribute":
        attr_key = cfg.get("attributeKey")
        if not attr_key or attr_key not in ctx.get("attributes", {}): return False, False
        
        ctx_val = ctx["attributes"][attr_key]
        op = cfg.get("attributeOp")
        cfg_val = cfg.get("attributeValue")
        
        actual = to_string(ctx_val)
        expected = to_string(cfg_val)

        if op == "eq": return (actual.lower().strip() == expected.lower().strip()), val
        if op == "neq": return (actual.lower().strip() != expected.lower().strip()), val
        if op == "contains":
            actual_l, expected_l = actual.lower().strip(), expected.lower().strip()
            if "," in actual:
                parts = [p.strip().lower() for p in actual.split(",")]
                if expected_l in parts: return True, val
            return (expected_l in actual_l), val
        if op in ["gt", "lt"]:
            try:
                a_f, e_f = float(actual), float(expected)
                if op == "gt": return a_f > e_f, val
                return a_f < e_f, val
            except: return False, False
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
        ctx["country"] = "ZZ"
        ctx["state"] = "MismatchState"
        return ctx

    if rtype == "user_list" and cfg.get("userIds"):
        ctx["userId"] = cfg["userIds"][0]
    elif rtype == "percentage":
        target_p = float(cfg.get("percentage", 0))
        for i in range(10000):
            uid = f"user-{i}"
            if get_bucket(flag_key, uid) < target_p:
                ctx["userId"] = uid
                break
    elif rtype == "geography":
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

    print(f"Running 200 tests with KEY-BASED evaluation...")

    for i in range(200):
        flag = random.choice(flags)
        rules = flag.get("rules", [])
        target_rule = random.choice(rules) if rules else {"type": "default"}
        rtype = target_rule.get("type")
        
        if rtype not in stats_types: stats_types[rtype] = {"passed": 0, "total": 0}
        stats_types[rtype]["total"] += 1

        context = generate_context_for_test(flag, target_rule, random.choice([True, False]))
        
        try:
            eval_url = f"{API_URL}/{flag.get('key')}/evaluate"
            resp = requests.post(eval_url, json=context, headers=HEADERS, timeout=2)
            if resp.status_code == 200:
                data = resp.json()
                actual_enabled = data.get("enabled")
                
                # Principal Sync: Capture server's evaluation time
                metadata = data.get("metadata", {})
                server_now_raw = metadata.get("evaluatedAt")
                server_now = None
                if server_now_raw:
                    server_now = datetime.fromisoformat(server_now_raw.replace("Z", "+00:00")).astimezone(timezone.utc)
                
                # Re-evaluate prediction using server's 'Now'
                expected_enabled, exp_reason = predict_evaluation(flag, context, server_now)
            else:
                actual_enabled = "ERR"
                expected_enabled = "FAIL"
                actual_reason = f"HTTP {resp.status_code}: {resp.text}"
        except Exception as e:
            actual_enabled = "ERR"
            expected_enabled = "FAIL"
            actual_reason = str(e)

        passed = (actual_enabled == expected_enabled)
        if passed:
            stats_total["passed"] += 1
            stats_types[rtype]["passed"] += 1
        else:
            stats_total["failed"] += 1
            print(f"FAIL | Key: {flag['key']:<15} | Rule: {rtype:<10} | Exp: {str(expected_enabled):<5} | Act: {str(actual_enabled):<5} | CtxID: {context.get('userId')}")

    # Summary
    print("\n" + "="*45)
    print(f"{'RULE TYPE':<15} | {'PASS':<6} | {'TOTAL':<6} | {'%'}")
    print("-" * 45)
    for rt, s in sorted(stats_types.items()):
        perc = (s['passed']/s['total'])*100 if s['total'] > 0 else 0
        print(f"{rt:<15} | {s['passed']:<6} | {s['total']:<6} | {perc:.1f}%")
    print("="*45)
    print(f"TOTAL: {stats_total['passed']}/200 ({(stats_total['passed']/200)*100:.1f}%)\n")

if __name__ == "__main__":
    main()
