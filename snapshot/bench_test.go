package snapshot

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethclient"
)


const (
  ipcpath = "/mnt/nvme0n1/ethereum/execution/Fullnode/geth.ipc"  
)
func cacheload(cachePath string) (*fastcache.Cache){
  
  //load cache 
  loadedCache, err := fastcache.LoadFromFile(cachePath) 
  if err != nil {
    log.Fatalf("Failed to load cache : %v", err)
  }

  return loadedCache
}
//DiskAccess 함수는 eth.getBalance를 호출
//disklayer에는 메모리DB를 적재
//디스크DB는 eth.getBlance로 역할 대체함
func DiskAccess(hash common.Hash){
  
}

func NewBlock(ipcpath string,number *big.Int) (map[common.Hash][]byte, error){

  client, _ := ethclient.Dial(ipcpath)
 
  block, err := client.BlockByNumber(context.Background(), number)
  if err != nil {
    return nil, err
  }

  accountMap := make(map[common.Hash][]byte)

  for _, tx := range block.Transactions() {
    to := tx.To()
    if to == nil {
      continue
    }

    balance, err := client.BalanceAt(context.Background(), *to, block.Number())
    if err != nil {
      return nil, fmt.Errorf("failed to get balance for account %s at block %d: %v", to.Hex(), number, err)
    }

    accountMap[common.BytesToHash(to.Bytes())] = balance.Bytes()
  }

  return accountMap, nil
}

func TestSnapshot2(t *testing.T){
  

 
  makeRoot := func(height uint64) common.Hash {
		var buffer [8]byte
		binary.BigEndian.PutUint64(buffer[:], height)
		return common.BytesToHash(buffer[:])
	}
  
  base := &diskLayer{
		diskdb: rawdb.NewMemoryDatabase(),
		root:   makeRoot(1),
		cache:  fastcache.New(1024 * 500),
	}
  
  snaps := &Tree{
    layers: map[common.Hash] snapshot {
      base.root: base,
    },
  }
  //big.NewInt(int64(i)
  diff_start  := int64(20474737)
  diff_end    := int64(20474737-128)
  
  
  // init diff layer
  accounts, err := NewBlock(ipcpath, big.NewInt(diff_end))
  if err != nil {
    t.Log("return block error : %v", err)
  }


  last := common.HexToHash("0x01")
  for i:= diff_end; i<=diff_start; i++ {
    head := common.BytesToHash([]byte{byte(i+2)})
    accounts, _ := NewBlock(ipcpath, big.NewInt(int64(i)))
    snaps.Update(head, last, nil, accounts, nil)
    last = head
    snaps.Cap(head, 128)
  }

  merged := (snaps.layers[last].(*diffLayer)).flatten().(*diffLayer)
  
   
  if have, want := len(merged.AccountList()), len(accounts); have != want {
		t.Errorf("AccountList() wrong: have %v, want %v", have, want)
	}

}
