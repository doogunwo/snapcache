package test

import (
	"encoding/csv"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
	"hash/fnv"

	"main/snapcache"
)

// FNV-1a 해시 함수 구현 (32비트)
func hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}


func Benchmark_SnapCache_Workload(b *testing.B) {
	// 캐시 초기화
	cache := snapcache.New[string, int](325300)
	// CSV 파일 미리 로드
	file, err := os.Open("dataset100.csv")
	if err != nil {
		b.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	// CSV 리더 초기화
	reader := csv.NewReader(file)

	// 헤더 무시
	_, err = reader.Read()
	if err != nil {
		b.Fatalf("Error reading CSV header: %v", err)
	}

	// 데이터를 미리 메모리로 로드
	var data [][]string
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		data = append(data, record)
	}

	// 고루틴 및 쓰레드 동기화를 위한 WaitGroup 초기화
	var wg sync.WaitGroup

	// 쓰레드 수 정의 (여기서는 10개의 고루틴)
	numThreads := 10

	// 벤치마크 반복을 위한 작업
	b.ResetTimer() // 타이머 리셋 (벤치마크 성능 측정에 불필요한 시간을 배제)
	for i := 0; i < b.N; i++ { // b.N만큼 반복
		totalRequests := 0
		cacheMisses := 0
		cacheHits := 0

		start := time.Now()

		// 고루틴을 통해 다중 스레드로 캐시 작업 수행
		for t := 0; t < numThreads; t++ {
			wg.Add(1)
			go func(threadID int) {
				defer wg.Done()

				for j := threadID; j < len(data); j += numThreads {
					// 데이터를 메모리에서 가져오기
					record := data[j]

					// 키와 값 가져오기
					key := record[0]
					valueStr := record[1]
					value, err := strconv.Atoi(valueStr) // 값을 정수로 변환
					if err != nil {
						b.Errorf("Invalid value in CSV: %s", valueStr)
						continue
					}

					hashedKey := strconv.Itoa(int(hashKey(key)))

					// 캐시 조회 및 업데이트
					_, ok := cache.Get(hashedKey)
					totalRequests++

					if !ok {
						// 캐시 미스 발생
						cacheMisses++
						cache.Set(hashedKey, value)
					} else {
						// 캐시 히트 발생
						cacheHits++
					}
				}
			}(t)
		}

		// 모든 고루틴이 끝날 때까지 대기
		wg.Wait()

		// 타이머 종료
		elapsed := time.Since(start)

		// QPS 계산
		qps := float64(totalRequests) / elapsed.Seconds()

		// 벤치마크 결과 출력
		b.Logf("Total Requests: %d", totalRequests)
		b.Logf("Cache Hits: %d", cacheHits)
		b.Logf("Cache Miss: %d", cacheMisses)
		b.Logf("Cache Hit Rate: %.2f%%", float64(cacheHits)/float64(totalRequests)*100)
		b.Logf("QPS (Queries Per Second): %.2f", qps)
		b.Logf("Elapsed Time: %s", elapsed)

		// 캐시 초기화
		cache.Purge()
	}
}
