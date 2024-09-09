import csv
import time
from cachemonCache import S3FIFO
def benchmark_s3fifo(cache, workload_file):
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
                cache.put(key, value)
            else:
                # Cache hit
                cache_hits += 1

            total_requests += 1

    elapsed = time.time() - start
    qps = total_requests / elapsed

    # Print benchmark results
    print(f"Total Requests: {total_requests}")
    print(f"Cache Hits: {cache_hits}")
    print(f"Cache Misses: {cache_misses}")
    print(f"Hit Rate: {cache_hits / total_requests * 100:.2f}%")
    print(f"Queries per Second (QPS): {qps:.2f}")

if __name__ == "__main__":

    # Create S3FIFO cache
    cache1 = S3FIFO(cache_size=32000)

    # Run benchmarks
    benchmark_s3fifo(cache1, "../test/dataset100.csv")
