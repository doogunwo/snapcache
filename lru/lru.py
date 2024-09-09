import csv
import time
from cachetools import LRUCache  # LRU 캐시를 위한 라이브러리

# LRU 캐시 벤치마크 함수
def benchmark_lru(cache, workload_file):
    total_requests = 0
    cache_hits = 0
    cache_misses = 0

    with open(workload_file, 'r') as file:
        reader = csv.reader(file)
        next(reader)  # Skip header

        start = time.time()

        for record in reader:
            key = record[0]
            value = 1

            if cache.get(key) is None:
                # Cache miss
                cache_misses += 1
                cache[key] = value  # 캐시 미스 시 값 추가
            else:
                # Cache hit
                cache_hits += 1

            total_requests += 1

    elapsed = time.time() - start
    qps = total_requests / elapsed

    # 벤치마크 결과 출력
    print(f"Total Requests: {total_requests}")
    print(f"Cache Hits: {cache_hits}")
    print(f"Cache Misses: {cache_misses}")
    print(f"Hit Rate: {cache_hits / total_requests * 100:.2f}%")
    print(f"Queries per Second (QPS): {qps:.2f}")

if __name__ == "__main__":
    workload_file = "../test/dataset100.csv"

    # LRU 캐시 생성 및 테스트
    cache = LRUCache(maxsize=32000)  # LRU 캐시 생성 (크기 512)
    print("LRU Cache Benchmark:")
    benchmark_lru(cache, workload_file)
