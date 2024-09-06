package snapcache

import (
	"encoding/csv"
	"os"
	"testing"
	"time"
)

const (
	workload1 = "../data/block.csv"
	workload2 = "../data/reuse_workload.csv"
)

// 벤치마크 함수 (SnapCache)
func Benchmark_SnapCache(b *testing.B) {
	cache := New[string, int](512) 
	runBenchmark(b, cache, workload1)
	cache2 := New[string, int](512)
	runBenchmark(b, cache2, workload2)
}

// 공통 벤치마크 로직 실행
func runBenchmark(b *testing.B, cache *SnapCache[string, int], workload string) {
	// 캐시 미스 및 캐시 히트 카운트
	totalRequests := 0
	cacheHits := 0
	cacheMisses := 0

	file, err := os.Open(workload)
	if err != nil {
		b.Fatalf("Error opening workload file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	_, err = reader.Read() // 헤더 읽기 (무시)
	if err != nil {
		b.Fatalf("Error reading CSV header: %v", err)
	}

	start := time.Now()

	for {
		record, err := reader.Read()
		if err != nil {
			break // 파일 끝에 도달하면 종료
		}

		key := record[0]
		value := 1
		_, ok := cache.Get(key)
		totalRequests++

		if !ok {
			//cache miss
			cacheMisses++
			cache.Set(key, value)

			if cache.currentSize >= cache.maxSize {
				cache.Evict() // 캐시가 가득 찼을 때 eviction 실행
			}
		} else {
			//cache hit
			cacheHits++

		}
	}

	// 타이머 종료
	elapsed := time.Since(start)
	qps := float64(totalRequests) / elapsed.Seconds()

	// 벤치마크 결과 출력
	b.Logf("Workload: %s", workload)
	b.Logf("Total Requests: %d", totalRequests)
	b.Logf("Cache Hits: %d", cacheHits)
	b.Logf("Cache Misses: %d", cacheMisses)
	b.Logf("Cache Hit Rate: %.2f%%", float64(cacheHits)/float64(totalRequests)*100)
	b.Logf("QPS (Queries Per Second): %.2f", qps)
	b.Logf("Elapsed Time: %s", elapsed)
	cache.Purge()
}
