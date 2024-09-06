package main

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	blockchainEntries = 1000000 // 생성할 블록체인 데이터셋 항목 수
	kvStoreEntries    = 5000000 // 생성할 KV Store 데이터셋 항목 수
	timeSeriesEntries = 10000000 // 생성할 타임 시리즈 데이터셋 항목 수
	fixedValue        = "fixed_value" // 모든 항목에 사용될 고정된 값
)

type datasetConfig struct {
	name        string
	entries     int
	keyGen      func() string
	value       string
	duplication float64 // 중복률 (0.0 ~ 1.0)
}

// Blockchain용 랜덤 키 생성
func generateBlockchainKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Key-Value Store용 랜덤 키 생성
func generateKVStoreKey() string {
	prefixes := []string{"user:", "txn:", "config:"}
	return prefixes[rand.Intn(len(prefixes))] + strconv.FormatUint(rand.Uint64(), 10)
}

// Time Series용 랜덤 키 생성 (타임스탬프 기반)
func generateTimeSeriesKey() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

// 데이터셋을 생성하는 함수
func generateDataset(config datasetConfig) {
	filename := fmt.Sprintf("%s_dataset.csv", config.name)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("파일 생성 오류 (%s): %v\n", config.name, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"key", "value"})
	if err != nil {
		fmt.Printf("CSV 헤더 작성 오류 (%s): %v\n", config.name, err)
		return
	}

	data := make(map[string]string)
	uniqueEntries := int(float64(config.entries) * (1 - config.duplication))

	// 중복되지 않은 데이터 생성
	for i := 0; i < uniqueEntries; i++ {
		key := config.keyGen()
		data[key] = config.value

		err := writer.Write([]string{key, config.value})
		if err != nil {
			fmt.Printf("CSV 행 작성 오류 (%s): %v\n", config.name, err)
			return
		}

		if i%50000 == 0 {
			fmt.Printf("%s: %d 엔트리 생성 완료\n", config.name, i)
		}
	}

	// 중복 데이터 생성
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	for i := uniqueEntries; i < config.entries; i++ {
		// 랜덤으로 중복 키 선택
		key := keys[rand.Intn(len(keys))]

		err := writer.Write([]string{key, config.value})
		if err != nil {
			fmt.Printf("CSV 행 작성 오류 (%s): %v\n", config.name, err)
			return
		}

		if i%50000 == 0 {
			fmt.Printf("%s: %d 엔트리 생성 완료\n", config.name, i)
		}
	}

	fmt.Printf("%s 데이터셋 생성 완료: %s\n", config.name, filename)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	configs := []datasetConfig{
		{
			name:        "blockchain",
			entries:     blockchainEntries,
			keyGen:      generateBlockchainKey,
			value:       fixedValue, // 고정된 값 사용
			duplication: 0.2, // 20% 중복
		},
		{
			name:        "kvstore",
			entries:     kvStoreEntries,
			keyGen:      generateKVStoreKey,
			value:       fixedValue, // 고정된 값 사용
			duplication: 0.3, // 30% 중복
		},
		{
			name:        "timeseries",
			entries:     timeSeriesEntries,
			keyGen:      generateTimeSeriesKey,
			value:       fixedValue, // 고정된 값 사용
			duplication: 0.1, // 10% 중복
		},
	}

	for _, config := range configs {
		generateDataset(config)
	}
}
