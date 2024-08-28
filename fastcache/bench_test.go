package main

import (
	"math/rand"
  	crand"crypto/rand"
	"testing"
	"time"
	"os"

	"github.com/VictoriaMetrics/fastcache"

	"runtime"
	"runtime/pprof"
)

type BenchmarkResult struct {
	Requests   int
	Hits       int
	Misses     int
	Hitrate    float64
	QPS        float64
	MemUsed    uint64
	WorkloadMem uint64
	CPUUsage   float64
}

func BenchmarkFastcache(b *testing.B) {

	f, err := os.Create("mem.prof")
	if err != nil {
		b.Fatalf("could not create memory profile: %v", err)
	}
	pprof.WriteHeapProfile(f)
	defer f.Close()

	// CPU 프로파일 시작
	fcpu, err := os.Create("cpu.prof")
	if err != nil {
		b.Fatalf("could not create CPU profile: %v", err)
	}
	pprof.StartCPUProfile(fcpu)
	defer pprof.StopCPUProfile()





	result := FastCache(1024*1024, 1000000, b)
  
  b.Logf("+-----------------+-----------------+\n")
  b.Logf("| Metric          | Value            |\n")
  b.Logf("+-----------------+-----------------+\n")
  b.Logf("| Requests        | %d               |\n", result.Requests)
  b.Logf("| Hits            | %d               |\n", result.Hits)
  b.Logf("| Misses          | %d               |\n", result.Misses)
  b.Logf("| Hit Rate        | %.2f%%           |\n", result.Hitrate*100)
  b.Logf("| QPS             | %.2f             |\n", result.QPS)
  b.Logf("| Memory Used     | %d MB            |\n", result.MemUsed/1024/1024)
  b.Logf("| Workload Memory | %d MB            |\n", result.WorkloadMem/1024/1024)
  b.Logf("| CPU Usage       | %.2f%%           |\n", result.CPUUsage)
  b.Logf("+-----------------+-----------------+\n")

}

func FastCache(cacheSize int, numKeys int, b *testing.B) BenchmarkResult {
	cache := fastcache.New(cacheSize)
	keys := generateKeys(numKeys)
	val := make([]byte, 128) // 128바이트의 임의 값

	var hits, misses int
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	startTime := time.Now()


	for i := 0; i < b.N; i++ {
		key := keys[rand.Intn(numKeys)]
		if _, found := cache.HasGet(nil, key[:]); found {
			hits++
		} else {
			misses++
			cache.Set(key, val)
		}
	}

	runtime.ReadMemStats(&memAfter)
	totalTime := time.Since(startTime).Seconds()
	totalRequests := hits + misses
	qps := float64(totalRequests) / totalTime
	hitRate := float64(hits) / float64(totalRequests)

	cpuUsage := CalculateCPUUsage()

	return BenchmarkResult{
		Requests:    totalRequests,
		Hits:        hits,
		Misses:      misses,
		Hitrate:     hitRate,
		QPS:         qps,
		MemUsed:     memAfter.HeapAlloc - memBefore.HeapAlloc,
		WorkloadMem: uint64(cacheSize) + uint64(memAfter.HeapAlloc-memBefore.HeapAlloc),
		CPUUsage:    cpuUsage,
	}
}

func generateKeys(num int) [][]byte {
	keys := make([][]byte, num)
	for i := 0; i < num; i++ {
		key := make([]byte, 32) // 32바이트의 임의 키 (Ethereum 키와 유사)
		crand.Read(key)
		keys[i] = key
	}
	return keys
}

func CalculateCPUUsage() float64 {
	// CPU 사용량을 계산하는 간단한 방법
	var cpuStats runtime.MemStats
	runtime.ReadMemStats(&cpuStats)
	cpuUsage := float64(cpuStats.PauseTotalNs) / float64(time.Second.Nanoseconds())
	return cpuUsage
}

