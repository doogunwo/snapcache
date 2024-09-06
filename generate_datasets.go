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
	blockchainEntries = 1000000
	kvStoreEntries    = 5000000
	timeSeriesEntries = 10000000
)

type datasetConfig struct {
	name        string
	entries     int
	keyGen      func() string
	valueGen    func() string
	duplication float64 // 중복률 (0.0 ~ 1.0)
}

func generateBlockchainKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateBlockchainValue() string {
	size := rand.Intn(3073) + 1024 // 1KB ~ 4KB
	bytes := make([]byte, size)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateKVStoreKey() string {
	prefixes := []string{"user:", "txn:", "config:"}
	return prefixes[rand.Intn(len(prefixes))] + strconv.FormatUint(rand.Uint64(), 10)
}

func generateKVStoreValue() string {
	size := rand.Intn(961) + 64 // 64B ~ 1KB
	bytes := make([]byte, size)
	rand.Read(bytes)
	return string(bytes)
}

func generateTimeSeriesKey() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func generateTimeSeriesValue() string {
	size := rand.Intn(385) + 128 // 128B ~ 512B
	bytes := make([]byte, size)
	rand.Read(bytes)
	return string(bytes)
}

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

	for i := 0; i < uniqueEntries; i++ {
		key := config.keyGen()
		value := config.valueGen()
		data[key] = value

		err := writer.Write([]string{key, value})
		if err != nil {
			fmt.Printf("CSV 행 작성 오류 (%s): %v\n", config.name, err)
			return
		}

		if i%100000 == 0 {
			fmt.Printf("%s: %d 엔트리 생성 완료\n", config.name, i)
		}
	}

	// 중복 데이터 생성
	for i := uniqueEntries; i < config.entries; i++ {
		var key string
		if len(data) > 0 {
			for k := range data {
				key = k
				break
			}
		} else {
			key = config.keyGen()
		}
		value := data[key]

		err := writer.Write([]string{key, value})
		if err != nil {
			fmt.Printf("CSV 행 작성 오류 (%s): %v\n", config.name, err)
			return
		}

		if i%100000 == 0 {
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
			valueGen:    generateBlockchainValue,
			duplication: 0.2, // 20% 중복
		},
		{
			name:        "kvstore",
			entries:     kvStoreEntries,
			keyGen:      generateKVStoreKey,
			valueGen:    generateKVStoreValue,
			duplication: 0.3, // 30% 중복
		},
		{
			name:        "timeseries",
			entries:     timeSeriesEntries,
			keyGen:      generateTimeSeriesKey,
			valueGen:    generateTimeSeriesValue,
			duplication: 0.1, // 10% 중복
		},
	}

	for _, config := range configs {
		generateDataset(config)
	}
}