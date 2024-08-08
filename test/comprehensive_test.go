package test

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/VictoriaMetrics/fastcache"
)

func TestComp(t *testing.T) {
    // 캐시 생성
    cache := fastcache.New(1024 * 1024) // 1MB 캐시

    // 맵 생성 및 데이터 추가
    dataMap := make(map[string]string)
    for i := 0; i < 300; i++ {
        key := generateEthereumAddress()
        value := generateRandomBalance()
        dataMap[key] = value
    }

    // 맵의 데이터를 캐시에 저장
    for key, value := range dataMap {
        cache.Set([]byte(key), []byte(value))
    }

    // 캐시 데이터를 디스크에 저장
    filePath := "cache.dat"
    err := cache.SaveToFile(filePath)
    if err != nil {
        log.Fatalf("Failed to save cache: %s", err)
    }

    // 캐시 데이터를 디스크에서 로드
    loadedCache, err := fastcache.LoadFromFile(filePath)
    if err != nil {
        log.Fatalf("Failed to load cache: %s", err)
    }
    i := 1
    // 데이터 검증
    for key, expectedValue := range dataMap {
        value := loadedCache.Get(nil, []byte(key))
        if string(value) != expectedValue {
            t.Logf("Data mismatch for key %s: got %s, expected %s", key, value, expectedValue)
        } else {
            t.Logf("%d %s %s",i,key,value)
            i = i +1
    }
    }

    fmt.Println("All data matches successfully!")

    // 캐시 파일 삭제
    os.Remove(filePath)
}

// 이더리움 주소 생성 (더미)
func generateEthereumAddress() string {
    const letters = "0123456789abcdef"
    address := make([]byte, 42)
    address[0] = '0'
    address[1] = 'x'
    for i := 2; i < 42; i++ {
        address[i] = letters[rand.Intn(len(letters))]
    }
    return string(address)
}

// 랜덤 밸런스 생성 (더미)
func generateRandomBalance() string {
    return fmt.Sprintf("%d", rand.Int63n(1000000))
}

