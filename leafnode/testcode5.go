package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/VictoriaMetrics/fastcache"
)

func main() {
	// 파일에서 캐시 로드
	cache, err := fastcache.LoadFromFile("account_cache.dat")
	if err != nil {
		fmt.Printf("파일에서 캐시 로드 실패: %v\n", err)
		return
	}

	// CSV 파일 열기
	file, err := os.Open("./account_frequencies.csv")
	if err != nil {
		fmt.Printf("CSV 파일 열기 실패: %v\n", err)
		return
	}
	defer file.Close()

	// CSV 파일 읽기
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("CSV 파일 읽기 실패: %v\n", err)
		return
	}

	// CSV 레코드 반복 처리, 헤더는 건너뜀, 캐시된 데이터 출력
	for _, record := range records[1:] {
		if len(record) < 2 {
			continue
		}

		address := record[0]
		cachedBalance := cache.Get(nil, []byte(address))
		if len(cachedBalance) > 0 {
			fmt.Printf("주소: %s, Balance: %s\n", address, cachedBalance)
		} else {
			fmt.Printf("주소: %s 캐시에서 찾을 수 없음\n", address)
		}
	}
}

