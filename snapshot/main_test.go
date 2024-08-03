package snapshot

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/big"
	"math/rand"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
  "github.com/ethereum/go-ethereum/ethdb/memorydb"
)



func randomHash() common.Hash {
	var hash common.Hash
	if n, err := crand.Read(hash[:]); n != common.HashLength || err != nil {
		panic(err)
	}
	return hash
}

// randomAccount generates a random account and returns it RLP encoded.
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

func emptyLayer() *diskLayer {
	return &diskLayer{
		diskdb: memorydb.New(),
		cache:  fastcache.New(500 * 1024),
	}
}

func randomStorageSet(accounts []string, hashes [][]string, nilStorage [][]string) map[common.Hash]map[common.Hash][]byte {
	storages := make(map[common.Hash]map[common.Hash][]byte)
	for index, account := range accounts {
		storages[common.HexToHash(account)] = make(map[common.Hash][]byte)

		if index < len(hashes) {
			hashes := hashes[index]
			for _, hash := range hashes {
				storages[common.HexToHash(account)][common.HexToHash(hash)] = randomHash().Bytes()
			}
		}
		if index < len(nilStorage) {
			nils := nilStorage[index]
			for _, hash := range nils {
				storages[common.HexToHash(account)][common.HexToHash(hash)] = nil
			}
		}
	}
	return storages
}

func randomBalance() uint64 {
	return rand.Uint64()
}

func generateRandomAccounts(n int) []string {
  accounts := make([]string,n)
  for i:=1; i<n; i++ {
    hash := randomHash()
    accounts[i] = hex.EncodeToString(hash[:])
  }
  return accounts
}


func generateRandomHashes(n int, m int) [][]string {
	hashes := make([][]string, n)
	for i := 0; i < n; i++ {
		hashes[i] = make([]string, m)
		for j := 0; j < m; j++ {
			hash := randomHash()
			hashes[i][j] = hex.EncodeToString(hash[:])
		}
	}
	return hashes
}


func NewBlock(tree *Tree, parentRoot common.Hash) *diffLayer {
  accounts := make(map[common.Hash]uint64)
  storage := make(map[common.Hash]map[common.Hash][]byte)
  for i := 0; i<300; i++ {
    account := common.BigToHash(big.NewInt(int64(i)))
    accounts[account] = randomBalance()

    storage[account] = make(map[common.Hash][]byte)
    for j := 0; j<5; j++ {
      slot := common.BigToHash(big.NewInt(int64(j)))
      storage[account][slot] = randomHash().Bytes()
    }
  }

  newLayer := newDiffLayer(emptyLayer(), common.Hash{}, make(map[common.Hash]struct{}), accounts, storage) 

  tree.layers[common.Hash{}] = newLayer 
  return newLayer
}

