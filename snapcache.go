package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/VictoriaMetrics/fastcache"
)

func main() {
	// Create cache
	cache := fastcache.New(1024 * 1024) // 1MB cache

	// Add dummy data to the cache
	for i := 0; i < 300; i++ {
		key := generateEthereumAddress()
		value := generateRandomBalance()
		cache.Set([]byte(key), []byte(value))
	}

	// Save cache to file
	filePath := "cache.dat"
	err := cache.SaveToFile(filePath)
	if err != nil {
		log.Fatalf("Failed to save cache: %s", err)
	}

	// Load cache from file
	loadedCache, err := fastcache.LoadFromFile(filePath)
	if err != nil {
		log.Fatalf("Failed to load cache: %s", err)
	}

	// Verify data is loaded correctly
	for i := 0; i < 300; i++ {
		key := generateEthereumAddress()
		v := loadedCache.Get(nil, []byte(key))
		if v != nil {
			fmt.Printf("Key: %s, Value: %s\n", key, v)
		}
	}
}

// Generate dummy Ethereum address
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

// Generate random balance
func generateRandomBalance() string {
	return fmt.Sprintf("%d", rand.Int63n(1000000))
}

