package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	ec "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
)


func main(){

  ipcpath := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth.ipc"
  dbpath := "/mnt/nvme0n1/ethereum/execution/Fullnode/geth/chaindata/ancient/chain"  
  client, err := ec.Dial(ipcpath)
  if err != nil {
    log.Fatalf("Failed to connect geth ipc ",err)
  }
  
  block,err := client.BlockByNumber(context.Background(),nil)
  if err != nil {
    log.Fatalf("block error : %v", err)
  }

  stateRoot := block.Root()

  ancientsDB, err := rawdb.NewLevelDBDatabaseWithFreezer(ancientsPath, 128, 1024, "", "", false)
  if err != nil {
    log.Fatal("dont open ancientDB : %v", err)
  }
  defer ancientsDB.Close()

  trieDB := triedb.NewDatabase(ancientsDB)
  stateTrie, err := trie.New(stateRoot, trieDB)

  if err != nil {
    log.Fatalf("dont open stateTrie : %v", err)
  }

   



}


