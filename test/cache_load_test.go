package test

import (
	"log"
  "testing"
	"github.com/VictoriaMetrics/fastcache"
)

func TestLoad (t *testing.T) {
	// Create cache

		// Save cache to file
	filePath := "cache.dat"


	// Load cache from file
	loadedCache, err := fastcache.LoadFromFile(filePath)
	if err != nil {
		log.Fatalf("Failed to load cache: %s", err)
	}
  
	// Verify data is loaded correctly
	for i := 0; i < 300; i++ {
		key := "AAA"
		v := loadedCache.Get(nil, []byte(key))
		if v != nil {
			t.Logf("Key: %s, Value: %s\n", key, v)
		}
	}
}

