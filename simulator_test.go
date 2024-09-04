package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
	"testing"
)


func datamaker(filename string){
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 헤더 작성
	writer.Write([]string{"key", "value"})

	// 총 데이터 수
	totalData := 51200
	oneHitWonderRatio := 0.60
	oneHitWonderCount := int(float64(totalData) * oneHitWonderRatio)
	// 랜덤 시드 설정
	rand.Seed(time.Now().UnixNano())

	// 데이터를 저장할 리스트
	var data [][]string

	// 1. 원히트원더 데이터 생성 (80%)
	for i := 1; i <= oneHitWonderCount; i++ {
		key := strconv.Itoa(i)
		value := strconv.Itoa(i)
		data = append(data, []string{key, value})
	}

	// 2. 자주 접근되는 데이터 생성 (20%)
	for i := oneHitWonderCount + 1; i <= totalData; i++ {
		key := strconv.Itoa(i)
		value := strconv.Itoa(i)
		// 자주 접근되므로 여러 번 추가
		accessCount := rand.Intn(10) + 1
		for j := 0; j < accessCount; j++ {
			data = append(data, []string{key, value})
		}
	}

	// 3. 데이터를 셔플 (순서 무작위화)
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})

	// 4. 셔플된 데이터를 CSV 파일에 기록
	for _, record := range data {
		writer.Write(record)
	}

	fmt.Println("Shuffled dataset with 80% one-hit-wonder data generated successfully.")
}

func Test_SnapCache_LoadAndTest(t *testing.T){
	
	
	cache := New[string, int](14364)
	file, err := os.Open("dataset.csv")
	if err != nil {
		t.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	// CSV 리더 초기화
	reader := csv.NewReader(file)

	// 헤더 무시
	_, err = reader.Read()
	if err != nil {
		t.Fatalf("Error reading CSV header: %v", err)
	}

	// QPS 및 캐시 미스 카운트 변수
	totalRequests := 0
	cacheMisses := 0
	cacheHits := 0

	start := time.Now()

	for i := 0; i<14000; i++{
		record, err := reader.Read()
		if err != nil {
			break // 파일의 끝에 도달하면 루프 종료
		}

		// 키와 값 가져오기
		key := record[0]
		valueStr := record[1]
		value, err := strconv.Atoi(valueStr) // 값을 정수로 변환

		if err != nil {
			t.Errorf("Invalid value in CSV: %s", valueStr)
			continue
		}

		hashedKey := strconv.Itoa(int(hashKey(key)))

		cache.Set(hashedKey, value)
		totalRequests++
	}

	for {
		record, err := reader.Read()
		if err != nil {
			break // 파일의 끝에 도달하면 루프 종료
		}

		key := record[0]
		hashedKey := strconv.Itoa(int(hashKey(key)))

		valueStr := record[1]
		value, err := strconv.Atoi(valueStr) // 값을 정수로 변환

		_, ok := cache.Get(hashedKey)
		totalRequests++

		if !ok {
			// cache miss
			cacheMisses++
			cache.Set(hashedKey, value)
		} else {
			// cache hits
			cacheHits++
		}
	}

	// 타이머 종료
	elapsed := time.Since(start)

	// QPS 계산
	qps := float64(totalRequests) / elapsed.Seconds()

	t.Logf("Total Requests: %d", totalRequests)
	t.Logf("Cache Hits: %d", cacheHits)
	t.Logf("Cache Miss: %d", cacheMisses)
	t.Logf("Cache Hit Rate: %.2f%%", float64(cacheHits)/float64(totalRequests)*100)
	t.Logf("QPS (Queries Per Second): %.2f", qps)
	t.Logf("Elapsed Time: %s", elapsed)
}

