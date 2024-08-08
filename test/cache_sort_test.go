package test

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

const (
	cacheSize = 1024 * 1024 * 1024 // 1GB
	numItems  = 1000000            // 1 million items
  key = "ABC"
  Value = "999999"

)

func randomHash() common.Hash {
	var hash common.Hash
	crand.Read(hash[:])
	return hash
}

func randomBalance() []byte {
	balance := uint256.NewInt(uint64(rand.Intn(1000000)))
	data, _ := rlp.EncodeToBytes(balance)
	if len(data) < 8 {
		paddedData := make([]byte, 8)
		copy(paddedData[8-len(data):], data)
		return paddedData
	}
	return data[:8]
}

type Account struct {
	Hash    common.Hash
	Balance []byte
}

func fillCache(cache *fastcache.Cache, accounts []Account) {
	for _, account := range accounts {
		cache.Set(account.Hash[:], account.Balance)
	}
  cache.Set([]byte(key),[]byte(Value))
}

func BenchmarkSortedCache(b *testing.B) {
	cache := fastcache.New(cacheSize)
	accounts := make([]Account, numItems)

	for i := 0; i < numItems; i++ {
		accounts[i] = Account{
			Hash:    randomHash(),
			Balance: randomBalance(),
		}
	}
  
  startsort := time.Now()
	sort.Slice(accounts, func(i, j int) bool {
		return binary.BigEndian.Uint64(accounts[i].Balance) < binary.BigEndian.Uint64(accounts[j].Balance)
	})
  sortDuration := time.Since(startsort)

	fillCache(cache, accounts)

	b.ResetTimer()
  startFind := time.Now()
	for i := 0; i < b.N; i++ {
    cache.Get(nil, []byte(key))
	}
  accessDuration := time.Since(startFind)
  b.StopTimer()
  b.Logf("sort time : %v",sortDuration)
  b.Logf("cache access time: %v",accessDuration)

}

func BenchmarkUnsortedCache(b *testing.B) {
	cache := fastcache.New(cacheSize)
	accounts := make([]Account, numItems)

	for i := 0; i < numItems; i++ {
		accounts[i] = Account{
			Hash:    randomHash(),
			Balance: randomBalance(),
		}
	}

	fillCache(cache, accounts)

	b.ResetTimer()

  startfind := time.Now()
	for i := 0; i < b.N; i++ {
    cache.Get(nil, []byte(key))
	}
  accessDuration := time.Since(startfind)
  b.StopTimer()
  b.Logf("cache access time :%s", accessDuration)
}


