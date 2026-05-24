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

def get_config(cfg, key):
    """Robust case-insensitive config lookup matching Go's getConfig."""
    if not cfg: return None
    if key in cfg: return cfg[key]
    target = key.lower()
    for k, v in cfg.items():
        if k.lower() == target:
            return v
    return None

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
    if not cfg: cfg = {}
    val = rule.get("value", False)
    user_id = ctx.get("userId", "")

    if rtype == "user_list":
        uids_raw = get_config(cfg, "userIds")
        if uids_raw is None: uids_raw = []
        if not isinstance(uids_raw, list): uids_raw = [uids_raw]
        uids = [to_string(uid).lower().strip() for uid in uids_raw]
        return (to_string(user_id).lower().strip() in uids), val
        
    elif rtype == "percentage":
        p = get_config(cfg, "percentage")
        if p is None or not user_id: return False, False
        bucket = get_bucket(flag_key, user_id)
        return bucket < float(p), val

    elif rtype == "gradual":
        if not user_id: return False, False
        try:
            start_at = datetime.fromisoformat(get_config(cfg, "startAt").replace("Z", "+00:00")).astimezone(timezone.utc)
            end_at = datetime.fromisoformat(get_config(cfg, "endAt").replace("Z", "+00:00")).astimezone(timezone.utc)
            now = server_now if server_now else datetime.now(timezone.utc)
            
            start_p = float(get_config(cfg, "startPercent") or 0)
            end_p = float(get_config(cfg, "endPercent") or 0)
            
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
            enable_at_raw = get_config(cfg, "enableAt")
            disable_at_raw = get_config(cfg, "disableAt")
            if enable_at_raw:
                ea = datetime.fromisoformat(enable_at_raw.replace("Z", "+00:00")).astimezone(timezone.utc)
                if now < ea: return False, False
            if disable_at_raw:
                da = datetime.fromisoformat(disable_at_raw.replace("Z", "+00:00")).astimezone(timezone.utc)
                if now > da: return False, False
            return True, val
        except: return False, False

    elif rtype == "geography":
        countries = get_config(cfg, "countries")
        states = get_config(cfg, "states")
        cities = get_config(cfg, "cities")
        zips = get_config(cfg, "zipCodes")
        if not any([countries, states, cities, zips]):
            return False, False

        if countries:
            if not any(c.lower() == ctx.get("country", "").lower().strip() for c in countries):
                return False, False
        if states:
            if not any(s.lower() == ctx.get("state", "").lower().strip() for s in states):
                return False, False
        if cities:
            if not any(c.lower() == ctx.get("city", "").lower().strip() for c in cities):
                return False, False
        if zips:
            if any(z.lower() == ctx.get("zipCode", "").lower().strip() for z in zips):
                return True, val
            return False, False
        return True, val

    elif rtype == "attribute":
        ak_raw = get_config(cfg, "attributeKey")
        if not ak_raw: return False, False
        
        # Case-insensitive attribute key lookup matching Go
        ctx_val = None
        for k, v in ctx.get("attributes", {}).items():
            if k.lower() == ak_raw.lower():
                ctx_val = v
                break
        
        if ctx_val is None: return False, False
        
        op = get_config(cfg, "attributeOp")
        cfg_val = get_config(cfg, "attributeValue")
        
        actual = to_string(ctx_val).lower().strip()
        expected = to_string(cfg_val).lower().strip()

        if op == "eq": return (actual == expected), val
        if op == "neq": return (actual != expected), val
        if op == "contains":
            if "," in actual:
                parts = [p.strip().lower() for p in actual.split(",")]
                if expected in parts: return True, val
            return (expected in actual), val
        if op in ["gt", "lt"]:
            try:
                a_f, e_f = float(actual), float(expected)
                if op == "gt": return a_f > e_f, val
                return a_f < e_f, val
            except:
                # Fallback to string comparison matching Go Principal Refinement
                if op == "gt": return actual > expected, val
                return actual < expected, val
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

    if rtype == "user_list":
        uids = get_config(cfg, "userIds")
        if uids: ctx["userId"] = uids[0]
    elif rtype == "percentage":
        target_p = float(get_config(cfg, "percentage") or 0)
        for i in range(10000):
            uid = f"user-{i}"
            if get_bucket(flag_key, uid) < target_p:
                ctx["userId"] = uid
                break
    elif rtype == "geography":
        c = get_config(cfg, "countries")
        s = get_config(cfg, "states")
        ci = get_config(cfg, "cities")
        z = get_config(cfg, "zipCodes")
        if c: ctx["country"] = c[0]
        if s: ctx["state"] = s[0]
        if ci: ctx["city"] = ci[0]
        if z: ctx["zipCode"] = z[0]
    elif rtype == "attribute":
        ak = get_config(cfg, "attributeKey")
        av = get_config(cfg, "attributeValue")
        if ak: ctx["attributes"][ak] = av or "pro"
        
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
                actual_reason = data.get("reason")
                
                metadata = data.get("metadata", {})
                server_now_raw = metadata.get("evaluatedAt")
                server_now = None
                if server_now_raw:
                    server_now = datetime.fromisoformat(server_now_raw.replace("Z", "+00:00")).astimezone(timezone.utc)
                
                expected_enabled, exp_reason = predict_evaluation(flag, context, server_now)
            else:
                actual_enabled = "ERR"
                expected_enabled = "FAIL"
                actual_reason = f"HTTP {resp.status_code}"
                exp_reason = "n/a"
        except Exception as e:
            actual_enabled = "ERR"
            expected_enabled = "FAIL"
            actual_reason = str(e)
            exp_reason = "n/a"

        passed = (actual_enabled == expected_enabled)
        if passed:
            stats_total["passed"] += 1
            stats_types[rtype]["passed"] += 1
        else:
            stats_total["failed"] += 1
            print(f"FAIL | Key: {flag['key']:<15} | Rule: {rtype:<10} | Exp: {str(expected_enabled):<5} | Act: {str(actual_enabled):<5} | ActReason: {actual_reason}")

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
