package lru

import (
	"encoding/csv"
	"os"
	"strconv"
	"testing"
	"time"
)

const (
	workload1 = "../data/block.csv"
	workload2 = "../data/reuse_workload.csv"
)

// 벤치마크 함수 (LRU Cache)
func Benchmark_LRU(b *testing.B) {
  
  cache := NewCache(nil)
  cache.SetCapacity(512)
  
  file, _ := os.Open(workload1)
  defer file.Close()

  reader := csv.NewReader(file)

  _, err = reader.Read()
  if err != nil {
    t.Fatalf(err)
  }

  total := 0
  miss :=  0
  hit := 0

  start := time.Now()

  for {
    record, _ := reader.Read()
    key := record[0]
    value := 1

    handle := cache.Get(0, key, nil)
    total++
    if handle 

  }

}


