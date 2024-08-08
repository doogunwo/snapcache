package snapshot

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
)





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



