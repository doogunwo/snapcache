package lru

import (
    "testing"
    "github.com/doogunwo/snapcache/data"
)

type releaserFunc struct {
	fn    func()
	value Value
}

func set(c *Cache, ns, key uint64, value Value, charge int, relf func()) *Handle {
	return c.Get(ns, key, func() (int, Value) {
		if relf != nil {
			return charge, releaserFunc{relf, value}
		}
		return charge, value
	})
}

// 문자열 키를 uint64로 변환하는 해시 함수
func hashStringToUint64(s string) uint64 {
    var h uint64
    for _, c := range s {
        h = h*31 + uint64(c)
    }
    return h
}

func BenchmarkLRUCache(b *testing.B) {
    // 캐시 크기를 512MB로 설정
    const cacheSize = 512 * 1024 * 1024 // 512MB
    const dataSize = 64                 // 각 키-값 쌍의 크기를 64바이트로 설정
    cache := NewCache(NewLRU(cacheSize))

    // 총 5120MB의 데이터를 생성
    const totalDataSize = cacheSize * 10 // 5120MB
    numEntries := totalDataSize / dataSize  // 데이터 크기 기반으로 삽입할 키-값 쌍의 개수 계산

    // 키-값 쌍 생성기
    generator := data.NewKeyValueGenerator(24, 24)  // 24바이트 크기의 키와 값 생성

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        key, value := generator.GenerateKeyValuePair()
        // uint64로 변환한 키를 사용
        uintKey := hashStringToUint64(key)
        
        // 캐시에 데이터 삽입 (적재)
        cache.Get(uintKey, func() (int, Value) {
            return len(value), value
        })

        // 캐시에서 키를 조회하여 적중률 확인
        handle := nsGetter.Get(uintKey, nil)
        if handle != nil {
            assert.Equal(b, value, handle.Value())
            handle.Release()
        }
    }
    b.StopTimer()  


    // 벤치마크가 끝난 후 캐시의 상태 출력
    b.Logf("캐시에 남아있는 항목 수: %d", cache.Nodes())
    stats := cache.GetStats()
    b.Logf("캐시 적중 수: %d, 캐시 미스 수: %d", stats.HitCount, stats.MissCount)
}

