package test

import (
	"log"
  "testing"
	"github.com/VictoriaMetrics/fastcache"
)

func TestSave(t *testing.T) {
	// Create cache
	cache := fastcache.New(1024 * 1024) // 1MB cache

	// Add dummy data to the cache
	for i := 0; i < 300; i++ {
		key := "AAA"
		value := "12222"
		cache.Set([]byte(key), []byte(value))
	}

	// Save cache to file
	filePath := "cache.dat"
	err := cache.SaveToFile(filePath)
	if err != nil {
		log.Fatalf("Failed to save cache: %s", err)
	}
  t.Log("cache creatr")
}


// Generate random balance
