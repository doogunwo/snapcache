package snapshot

import (
  crand"crypto/rand"
  "math/rand"
	"testing"
  "path/filepath"
  "encoding/binary"
  "fmt"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
  "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

const (
	cacheSize = 1024 * 1024 * 1024 // 1GB
	numItems  = 1000000            // 1백만 항목
)

func randomHash() common.Hash {
	var hash common.Hash
	crand.Read(hash[:])
	return hash
}


func randomAccount() []byte {
	a := &types.StateAccount{
		Balance:  uint256.NewInt(rand.Uint64()),
		Nonce:    rand.Uint64(),
		Root:     randomHash(),
		CodeHash: types.EmptyCodeHash[:],
	}
	data, _ := rlp.EncodeToBytes(a)
	return data
}

func randomBalance() []byte {
	balance := uint256.NewInt(uint64(rand.Intn(1000000))) // 최대 10^6 밸런스
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
}

func emptyLayer() *diskLayer {
  dataDir := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth/chaindata"
  ancientDir := filepath.Join(dataDir, "ancient/chain")

  options := rawdb.OpenOptions{
        Type:              "pebble",
        Directory:         dataDir,
        AncientsDirectory: ancientDir,
        Namespace:         "namespace",
        Cache:             128,
        Handles:           128,
        ReadOnly:          false,
        Ephemeral:         false,
    }

  db, err := rawdb.Open(options)

  if err != nil {panic("Failed to open database: " + err.Error())}
  return &diskLayer{
        diskdb: db,
        cache:  fastcache.New(5 * 1024 * 1024 ),
        root:   common.HexToHash("0x01"),
    }
}

func NewBlock() 

func TestSnapshot(t *testing.T){

  base := emptyLayer()

  snaps := &Tree{
    layers : map[common.Hash] snapshot {
      base.root : base,
    },
  }
  
  setAccount := func(accKey string) map[common.Hash][]byte {
		return map[common.Hash][]byte{
			common.HexToHash(accKey): randomAccount(),
		}
	}
  
  makeRoot := func(height uint64) common.Hash {
		var buffer [8]byte
		binary.BigEndian.PutUint64(buffer[:], height)
		return common.BytesToHash(buffer[:])
	}

  var (
    last = common.HexToHash("0x01")
    head common.Hash
  )

  for i := 0; i < 129; i++ {
		head = makeRoot(uint64(i + 2))
		snaps.Update(head, last, nil, setAccount(fmt.Sprintf("%d", i+2)), nil)
		last = head
		snaps.Cap(head, 128) // 130 layers (128 diffs + 1 accumulator + 1 disk)
	}

 



}



