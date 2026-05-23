import asyncio
import aiohttp
import time
import random
import sys
import math
from datetime import datetime

# --- CONFIGURATION ---
API_URL = "http://localhost:8081/api/flags"
API_KEY = "test-api-key"
HEADERS = {"X-API-KEY": API_KEY}

# Extreme targets
TOTAL_REQUESTS = 1000000 
CONCURRENCY_LIMIT = 5000  # Simultaneous open connections
TIMEOUT_SECONDS = 30
# ---------------------

# WARNINGS:
# - This test is designed to push the limits of the API and may cause instability.
# - Ensure the API server is configured to handle this load and monitor resources closely.

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
        reason_str = str(reason)
        if "Event loop is closed" in reason_str: reason_str = "Loop Closed"
        self.error_reasons[reason_str] = self.error_reasons.get(reason_str, 0) + 1

    def report_progress(self):
        now = time.perf_counter()
        interval = now - self.last_report
        
        if interval >= 1.0:
            current_rps = (self.success - self.last_success) / interval
            sys.stdout.write(
                f"\r[{datetime.now().strftime('%H:%M:%S')}] "
                f"Progress: {self.success + self.errors:,} | "
                f"Success: {self.success:,} | "
                f"Errors: {self.errors:,} | "
                f"Current RPS: {current_rps:.0f}"
            )
            sys.stdout.flush()
            self.last_report = now
            self.last_success = self.success

def get_percentile(data, percentile):
    if not data: return 0
    size = len(data)
    sorted_data = sorted(data)
    index = (size - 1) * percentile
    lower = math.floor(index)
    upper = math.ceil(index)
    if lower == upper:
        return sorted_data[int(index)]
    d0 = sorted_data[lower] * (upper - index)
    d1 = sorted_data[upper] * (index - lower)
    return d0 + d1

async def worker(session, flag_ids, stats):
    """Continuously slams API until TOTAL_REQUESTS reached."""
    while (stats.success + stats.errors) < TOTAL_REQUESTS:
        fid = random.choice(flag_ids)
        ctx = {"userId": f"stress-{random.randint(1, 1000000)}", "attributes": {"load": "max"}}
        
        try:
            start = time.perf_counter()
            async with session.post(
                f"{API_URL}/{fid}/evaluate", 
                json=ctx, 
                headers=HEADERS, 
                timeout=aiohttp.ClientTimeout(total=TIMEOUT_SECONDS)
            ) as resp:
                if resp.status == 200:
                    await resp.read() 
                    stats.success += 1
                    stats.latencies.append(time.perf_counter() - start)
                else:
                    stats.add_error(f"HTTP {resp.status}")
        except Exception as e:
            stats.add_error(f"{type(e).__name__}: {str(e)}")
            await asyncio.sleep(0.01) # Short backoff
        
        stats.report_progress()

async def main():
    print(f"--- EXTREME CONCURRENCY TEST ---")
    print(f"Target: {API_URL}")
    print(f"Concurrency: {CONCURRENCY_LIMIT}")
    print(f"Max Requests: {TOTAL_REQUESTS:,}")
    print("-" * 35)

    connector = aiohttp.TCPConnector(limit=CONCURRENCY_LIMIT, limit_per_host=CONCURRENCY_LIMIT)
    async with aiohttp.ClientSession(connector=connector) as session:
        # Get flags once
        try:
            async with session.get(API_URL, headers=HEADERS) as resp:
                if resp.status != 200:
                    print(f"Fetch flags fail: {resp.status}")
                    return
                flags = await resp.json()
                flag_ids = [f.get("id") or f.get("_id") for f in flags if f.get("enabled")]
        except Exception as e:
            print(f"Initial fetch error: {e}")
            return

        if not flag_ids:
            print("No enabled flags found.")
            return

        stats = StressStats()
        print(f"Starting workers with {len(flag_ids)} enabled flags...")
        
        tasks = [worker(session, flag_ids, stats) for _ in range(CONCURRENCY_LIMIT)]
        await asyncio.gather(*tasks)

    total_time = time.perf_counter() - stats.start_time
    count = len(stats.latencies)
    avg_lat = (sum(stats.latencies) / count) * 1000 if count else 0
    p95 = get_percentile(stats.latencies, 0.95) * 1000
    p99 = get_percentile(stats.latencies, 0.99) * 1000
    
    print("\n\n" + "="*40)
    print(f"FINAL STRESS TEST REPORT")
    print("-" * 40)
    print(f"Total Time:      {total_time:.2f}s")
    print(f"Total Requests:  {stats.success + stats.errors:,}")
    print(f"Success Rate:    {(stats.success/(stats.success+stats.errors or 1))*100:.1f}%")
    print(f"Avg RPS:         {(stats.success + stats.errors)/total_time:.0f}")
    print(f"Avg Latency:     {avg_lat:.2f}ms")
    print(f"P95 Latency:     {p95:.2f}ms")
    print(f"P99 Latency:     {p99:.2f}ms")

    if stats.error_reasons:
        print("\nTop Error Reasons:")
        for reason, count in sorted(stats.error_reasons.items(), key=lambda x: x[1], reverse=True)[:5]:
            print(f"- {reason}: {count:,}")
    print("="*40)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nTest aborted.")

