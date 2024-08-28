package main

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/VictoriaMetrics/fastcache"
	"go.uber.org/goleak"
)

func generateValue(size int) []byte {
	value := make([]byte, size)
	rand.Read(value)
	return value
}

func benchmarkFastCache(b *testing.B, valueSize int) {
	cache := fastcache.New(10 * 1024 * 1024) // 10MB 캐시 할당
	key := []byte("fixed_key")              // 고정된 크기의 키

	for i := 0; i < b.N; i++ {
		value := generateValue(valueSize)
		cache.Set(key, value)
		cache.Get(nil, key)
        b.Log(key,value)
	}
}

func TestFastCacheBenchmark(t *testing.T) {
	valueSizes := []int{150, 160, 180, 200, 256} // 다양한 크기의 값

	for _, size := range valueSizes {
		b := testing.Benchmark(func(b *testing.B) {
			benchmarkFastCache(b, size)
		})
		fmt.Printf("Value Size: %d bytes, Time per operation: %s\n", size, b.T)
	}
}

func TestMain(m *testing.M) {
	// 고루틴 누수 감지
    goleak.VerifyTestMain(m)
	m.Run()
}


