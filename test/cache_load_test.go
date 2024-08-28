package test

import (
	"bufio"
	"log"
	"os"
	"testing"
  "math/big"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestLoad (t *testing.T) {
	// Create cache

		// Save cache to file
	filePath := "addrOver100"


	// Load cache from file
	loadedCache, err := fastcache.LoadFromFile(filePath)
	if err != nil {
		log.Fatalf("Failed to load cache: %s", err)
	}
  
  txtfild := "addrOver100.txt"
  file, _ := os.Open(txtfild)
  scanner := bufio.NewScanner(file)
  cnt := 1
  for scanner.Scan() {
    line := scanner.Text()
    address := common.HexToAddress(line)
    accountHash := crypto.Keccak256Hash(address.Bytes())
    data, found := loadedCache.HasGet(nil, accountHash.Bytes())
    if found {
      balance := new(big.Int).SetBytes(data)
      t.Logf("%d: %s", cnt, balance.String())
    } else {
      t.Log(cnt," dont found")
    }

    cnt = cnt +1
  }
}

