package main

import (
	"bufio"
	crand "crypto/rand"

	"math/big"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
  "runtime"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

)

type Account struct {
	Nonce    uint64
	Balance  *big.Int
	Root     common.Hash // storage root
	CodeHash []byte
}

func generateRandomAccount() []byte {

	randGen := rand.New(rand.NewSource(1))

	nonce := uint64(0)
	balance := big.NewInt(0).Rand(randGen, big.NewInt(1e18))
	root := common.BytesToHash(make([]byte, 32))                   // Empty storage root (for simplicity)
	codeHash := crypto.Keccak256Hash([]byte("randomCode")).Bytes() // Random code hash

  account := Account{
		Nonce:    nonce,
		Balance:  balance,
		Root:     root,
    CodeHash: codeHash[:],
	}

  encodedBlob, _ := rlp.EncodeToBytes(account)
  return encodedBlob

}

func TestFastCache(t *testing.T) {

  cache := fastcache.New(13 * 1024  * 1024)
  
  var (
    cacheMisses = uint64(0)
    cacheHit    = uint64(0)
  )

  file, err := os.Open("../addr_fullnode2.txt") // key load
  if err != nil {
    t.Logf("txt error : %v", err)
  }

  var memStats runtime.memStats
  runtime.ReadMemStats(&memStats)
  startGC := memStats.NumGC

  b.ResetTimer()

  cnt := 1
  scanner := bufio.NewScanner(file)
  for scanner.Scan(){
    line := scanner.Text()
    address := common.HexToAddress(line)
    acc := crypto.Keccak256Hash(address.Bytes())
    
    _, found := cache.HasGet(nil, acc[:])
    if found {
      atomic.AddUint64(&cacheHit, 1)
    } else {
      atomic.AddUint64(&cacheMisses, 1)
      blob := generateRandomAccount()
      cache.Set(acc[:], blob)
    }
	cnt = cnt + 1
  }

  t.Log("complete")
  t.Log("Total execution:",cnt," cachehits : " ,cacheHit," cacheMisses", cacheMisses)
}

