import asyncio
import aiohttp
import time
import random
import sys
import math
import orjson
from datetime import datetime

# --- CONFIGURATION ---
API_URL = "http://localhost:8081/api/flags"
API_KEY = "test-api-key"
HEADERS = {
    "X-API-KEY": API_KEY,
    "Content-Type": "application/json",
    "Connection": "keep-alive"
}

TOTAL_REQUESTS = 10000 
CONCURRENCY_LIMIT = 500
TIMEOUT_SECONDS = 10
# ---------------------

class StressStats:
    def __init__(self):
        self.success = 0
        self.errors = 0
        self.error_reasons = {}
        self.latencies = []
        self.start_time = time.perf_counter()
        self.last_report = time.perf_counter()
        self.last_success = 0

    def add_error(self, reason):
        self.errors += 1
        reason_str = str(reason)[:50]
        self.error_reasons[reason_str] = self.error_reasons.get(reason_str, 0) + 1

    def report_progress(self, force=False):
        now = time.perf_counter()
        interval = now - self.last_report
        
        if interval >= 1.0 or force:
            current_rps = (self.success - self.last_success) / interval
            sys.stdout.write(
                f"\r[{datetime.now().strftime('%H:%M:%S')}] "
                f"Done: {self.success + self.errors:,} | "
                f"RPS: {current_rps:.0f} | "
                f"Err: {self.errors:,}"
            )
            sys.stdout.flush()
            self.last_report = now
            self.last_success = self.success

async def worker(session, flag_ids, stats, payloads):
    """Zero-allocation-ish loop for maximum throughput."""
    url_template = f"{API_URL}/{{}}/evaluate"
    
    # Pre-select URLs to avoid f-string overhead in loop
    urls = [url_template.format(fid) for fid in flag_ids]
    
    while (stats.success + stats.errors) < TOTAL_REQUESTS:
        url = random.choice(urls)
        # Use pre-encoded random payload
        payload = random.choice(payloads)
        
        try:
            start = time.perf_counter()
            async with session.post(url, data=payload, timeout=TIMEOUT_SECONDS) as resp:
                if resp.status == 200:
                    await resp.release() # Faster than read() if we don't need body
                    stats.success += 1
                    # Only track latency subset to save memory/CPU if needed, but 1M is fine for now
                    stats.latencies.append(time.perf_counter() - start)
                else:
                    stats.add_error(f"HTTP {resp.status}")
        except Exception as e:
            stats.add_error(type(e).__name__)
        
        # Report less frequently in worker to save CPU
        if stats.success % 100 == 0:
            stats.report_progress()

def get_percentile(data, percentile):
    if not data: return 0
    size = len(data)
    sorted_data = sorted(data)
    index = (size - 1) * percentile
    lower = math.floor(index)
    upper = math.ceil(index)
    if lower == upper:
        return sorted_data[int(index)]
    return sorted_data[lower] * (upper - index) + sorted_data[upper] * (index - lower)

async def main():
    print(f"--- OPTIMIZED STRESS TEST ---")
    print(f"Concurrency: {CONCURRENCY_LIMIT} | Target: {TOTAL_REQUESTS:,}")
    
    # Pre-encode 1000 random payloads to avoid JSON overhead in loop
    print("Pre-encoding payloads...")
    payloads = [
        orjson.dumps({
            "userId": f"u-{random.randint(1, 1000000)}", 
            "attributes": {"tier": "gold", "v": str(i)}
        }) for i in range(1000)
    ]

    connector = aiohttp.TCPConnector(
        limit=CONCURRENCY_LIMIT, 
        limit_per_host=CONCURRENCY_LIMIT,
        ttl_dns_cache=300,
        use_dns_cache=True
    )
    
    async with aiohttp.ClientSession(connector=connector, headers=HEADERS) as session:
        try:
            async with session.get(API_URL) as resp:
                flags = await resp.json()
                flag_keys = [f["key"] for f in flags if f.get("enabled")]
        except Exception as e:
            print(f"Fetch fail: {e}")
            return

        if not flag_keys:
            print("No flags.")
            return

        stats = StressStats()
        print(f"Slamming with {len(flag_keys)} flags...")
        
        tasks = [worker(session, flag_keys, stats, payloads) for _ in range(CONCURRENCY_LIMIT)]
        await asyncio.gather(*tasks)
        stats.report_progress(force=True)

    total_time = time.perf_counter() - stats.start_time
    count = len(stats.latencies)
    avg_lat = (sum(stats.latencies) / count) * 1000 if count else 0
    p95 = get_percentile(stats.latencies, 0.95) * 1000
    
    print(f"\n\nFINAL: {count/total_time:.0f} RPS | Avg: {avg_lat:.2f}ms | P95: {p95:.2f}ms")

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nAborted.")
