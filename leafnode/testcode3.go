package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
)

func main() {
	// 노드에 연결
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// 최신 블록 헤더 가져오기
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to get the latest block header: %v", err)
	}

	// 상태 트리의 루트 해시 가져오기
	stateRoot := header.Root

	// 디스크 데이터베이스 열기
	diskdb, err := rawdb.NewLevelDBDatabase("path/to/chaindata", 128, 1024, "", true)
	if err != nil {
		log.Fatalf("Failed to open LevelDB database: %v", err)
	}
	defer diskdb.Close()

	// 트리 데이터베이스 생성
	trieDb := triedb.NewDatabase(diskdb,nil)

	// 상태 트리 생성
	stateTrie, err := trie.NewStateTrie(trie.TrieID(stateRoot), trieDb)
	if err != nil {
		log.Fatalf("Failed to create state trie: %v", err)
	}

	// 리프 노드 탐색 및 출력
	it, err := stateTrie.NodeIterator(nil)
  if err != nil {
    log.Fatalf("stateTrie, err : %v",err)
  }
	for it.Next(true) {
		if it.Leaf() {
			key := it.Path()
			value := it.LeafBlob()
			fmt.Printf("Key: %x, Value: %x\n", key, value)
		}
	}

	if it.Error() != nil {
		log.Fatalf("Iterator error: %v", it.Error())
	}
}

